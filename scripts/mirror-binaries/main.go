package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Manifest represents the structure of a runtime manifest
type Manifest struct {
	Versions map[string]map[string]*Download `json:"versions"`
}

// Download represents a single download entry
type Download struct {
	URL    string `json:"url"`
	SHA256 string `json:"sha256,omitempty"`
}

// MirrorJob represents a single file to mirror
type MirrorJob struct {
	Runtime  string
	Version  string
	Platform string
	URL      string
	SHA256   string
	R2Key    string
}

// Stats tracks mirroring statistics
type Stats struct {
	Total     int64
	Skipped   int64
	Mirrored  int64
	Failed    int64
	BytesDown int64
}

var (
	runtime      = flag.String("runtime", "", "Runtime to mirror (node, python, ruby, or all)")
	dryRun       = flag.Bool("dry-run", false, "Report what would be done without doing it")
	syncOnly     = flag.Bool("sync-only", false, "Only mirror files not already in R2")
	manifestDir  = flag.String("manifest-dir", "src/internal/manifest/data", "Directory containing manifest files")
	r2Endpoint   = flag.String("r2-endpoint", "", "R2 endpoint URL")
	r2Bucket     = flag.String("r2-bucket", "", "R2 bucket name")
	r2AccessKey  = flag.String("r2-access-key", "", "R2 access key ID")
	r2SecretKey  = flag.String("r2-secret-key", "", "R2 secret access key")
	workers      = flag.Int("workers", 10, "Number of parallel workers")
	retries      = flag.Int("retries", 3, "Number of retries for failed downloads")
)

func main() {
	flag.Parse()

	if *runtime == "" {
		fmt.Fprintln(os.Stderr, "Error: --runtime is required (node, python, ruby, or all)")
		os.Exit(1)
	}

	if !*dryRun {
		if *r2Endpoint == "" || *r2Bucket == "" || *r2AccessKey == "" || *r2SecretKey == "" {
			fmt.Fprintln(os.Stderr, "Error: R2 credentials required (--r2-endpoint, --r2-bucket, --r2-access-key, --r2-secret-key)")
			os.Exit(1)
		}
	}

	runtimes := []string{*runtime}
	if *runtime == "all" {
		runtimes = []string{"node", "python", "ruby"}
	}

	// Initialize S3 client for R2
	var s3Client *s3.Client
	var existingKeys map[string]bool

	if !*dryRun {
		var err error
		s3Client, err = createS3Client()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating S3 client: %v\n", err)
			os.Exit(1)
		}

		if *syncOnly {
			fmt.Println("Fetching existing files from R2...")
			existingKeys, err = listExistingKeys(s3Client)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error listing R2 contents: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Found %d existing files in R2\n", len(existingKeys))
		}
	}

	// Collect all jobs
	var jobs []MirrorJob
	for _, rt := range runtimes {
		manifestPath := filepath.Join(*manifestDir, rt+".json")
		rtJobs, err := loadJobs(rt, manifestPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading manifest for %s: %v\n", rt, err)
			os.Exit(1)
		}
		jobs = append(jobs, rtJobs...)
	}

	fmt.Printf("Total jobs to process: %d\n", len(jobs))

	if *dryRun {
		fmt.Println("\n[DRY RUN] Would mirror the following files:")
		for _, job := range jobs {
			fmt.Printf("  %s -> %s\n", job.URL, job.R2Key)
		}
		fmt.Printf("\nTotal: %d files\n", len(jobs))
		return
	}

	// Filter jobs if sync-only
	if *syncOnly && existingKeys != nil {
		var filtered []MirrorJob
		for _, job := range jobs {
			if !existingKeys[job.R2Key] {
				filtered = append(filtered, job)
			}
		}
		skipped := len(jobs) - len(filtered)
		fmt.Printf("Skipping %d files already in R2, %d remaining\n", skipped, len(filtered))
		jobs = filtered
	}

	if len(jobs) == 0 {
		fmt.Println("No files to mirror")
		return
	}

	// Process jobs with worker pool
	stats := processJobs(s3Client, jobs)

	// Print summary
	fmt.Println("\n=== Mirror Summary ===")
	fmt.Printf("Total:    %d\n", stats.Total)
	fmt.Printf("Mirrored: %d\n", stats.Mirrored)
	fmt.Printf("Skipped:  %d\n", stats.Skipped)
	fmt.Printf("Failed:   %d\n", stats.Failed)
	fmt.Printf("Bytes:    %d MB\n", stats.BytesDown/(1024*1024))

	if stats.Failed > 0 {
		os.Exit(1)
	}
}

func createS3Client() (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			*r2AccessKey,
			*r2SecretKey,
			"",
		)),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(*r2Endpoint)
	})

	return client, nil
}

func listExistingKeys(client *s3.Client) (map[string]bool, error) {
	keys := make(map[string]bool)
	paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
		Bucket: r2Bucket,
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return nil, err
		}
		for _, obj := range page.Contents {
			keys[*obj.Key] = true
		}
	}

	return keys, nil
}

func loadJobs(runtime, manifestPath string) ([]MirrorJob, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	var jobs []MirrorJob
	for version, platforms := range manifest.Versions {
		for platform, dl := range platforms {
			if dl == nil || dl.URL == "" {
				continue
			}

			// Determine file extension from URL
			ext := getExtension(dl.URL)
			r2Key := fmt.Sprintf("%s/%s/%s%s", runtime, version, platform, ext)

			jobs = append(jobs, MirrorJob{
				Runtime:  runtime,
				Version:  version,
				Platform: platform,
				URL:      dl.URL,
				SHA256:   dl.SHA256,
				R2Key:    r2Key,
			})
		}
	}

	return jobs, nil
}

func getExtension(url string) string {
	// Handle common archive extensions
	if strings.HasSuffix(url, ".tar.gz") {
		return ".tar.gz"
	}
	if strings.HasSuffix(url, ".tar.xz") {
		return ".tar.xz"
	}
	if strings.HasSuffix(url, ".tar.bz2") {
		return ".tar.bz2"
	}
	if strings.HasSuffix(url, ".zip") {
		return ".zip"
	}
	if strings.HasSuffix(url, ".7z") {
		return ".7z"
	}
	// Fallback: extract from URL
	base := filepath.Base(url)
	if idx := strings.Index(base, "."); idx != -1 {
		return base[idx:]
	}
	return ""
}

func processJobs(client *s3.Client, jobs []MirrorJob) *Stats {
	stats := &Stats{Total: int64(len(jobs))}
	jobChan := make(chan MirrorJob, len(jobs))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobChan {
				if err := mirrorFile(client, job, stats); err != nil {
					fmt.Fprintf(os.Stderr, "Error mirroring %s: %v\n", job.R2Key, err)
					atomic.AddInt64(&stats.Failed, 1)
				} else {
					atomic.AddInt64(&stats.Mirrored, 1)
				}
			}
		}()
	}

	// Queue jobs
	for _, job := range jobs {
		jobChan <- job
	}
	close(jobChan)

	wg.Wait()
	return stats
}

func mirrorFile(client *s3.Client, job MirrorJob, stats *Stats) error {
	var lastErr error

	for attempt := 1; attempt <= *retries; attempt++ {
		if attempt > 1 {
			fmt.Printf("Retry %d/%d for %s\n", attempt, *retries, job.R2Key)
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		err := doMirror(client, job, stats)
		if err == nil {
			return nil
		}
		lastErr = err
	}

	return lastErr
}

func doMirror(client *s3.Client, job MirrorJob, stats *Stats) error {
	// Download file
	httpClient := &http.Client{Timeout: 10 * time.Minute}
	resp, err := httpClient.Get(job.URL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	// Read body and calculate checksum
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read failed: %w", err)
	}

	atomic.AddInt64(&stats.BytesDown, int64(len(body)))

	// Verify checksum if provided
	if job.SHA256 != "" {
		hash := sha256.Sum256(body)
		actual := hex.EncodeToString(hash[:])
		if actual != job.SHA256 {
			return fmt.Errorf("checksum mismatch: expected %s, got %s", job.SHA256, actual)
		}
	}

	// Determine content type
	contentType := "application/octet-stream"
	if strings.HasSuffix(job.R2Key, ".tar.gz") {
		contentType = "application/gzip"
	} else if strings.HasSuffix(job.R2Key, ".zip") {
		contentType = "application/zip"
	} else if strings.HasSuffix(job.R2Key, ".tar.xz") {
		contentType = "application/x-xz"
	}

	// Upload to R2
	_, err = client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:       r2Bucket,
		Key:          aws.String(job.R2Key),
		Body:         strings.NewReader(string(body)),
		ContentType:  aws.String(contentType),
		CacheControl: aws.String("public, max-age=31536000, immutable"),
	})
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	fmt.Printf("Mirrored: %s (%d bytes)\n", job.R2Key, len(body))
	return nil
}
