package core

// Thumbnailer bounds the number of concurrent thumbnail generations so a
// folder full of images doesn't spawn unbounded decode/resize work at once
// (spec §3: "worker pools for...thumbnail generation").
type Thumbnailer struct {
	jobs chan thumbnailJob
}

type thumbnailJob struct {
	cacheDir, srcPath string
	maxSize           int
	result            chan thumbnailResult
}

type thumbnailResult struct {
	path string
	err  error
}

// NewThumbnailer starts workers goroutines processing thumbnail requests.
func NewThumbnailer(workers int) *Thumbnailer {
	t := &Thumbnailer{jobs: make(chan thumbnailJob)}
	for i := 0; i < workers; i++ {
		go t.worker()
	}
	return t
}

func (t *Thumbnailer) worker() {
	for job := range t.jobs {
		path, err := CacheThumbnail(job.cacheDir, job.srcPath, job.maxSize)
		job.result <- thumbnailResult{path: path, err: err}
	}
}

// Generate requests a thumbnail for srcPath and blocks until a worker
// completes it.
func (t *Thumbnailer) Generate(cacheDir, srcPath string, maxSize int) (string, error) {
	result := make(chan thumbnailResult, 1)
	t.jobs <- thumbnailJob{cacheDir: cacheDir, srcPath: srcPath, maxSize: maxSize, result: result}
	r := <-result
	return r.path, r.err
}

// Close stops all workers. Callers must not call Generate after Close.
func (t *Thumbnailer) Close() {
	close(t.jobs)
}
