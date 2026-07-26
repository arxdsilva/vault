// Builds the URL that streams path's raw bytes from the Go fileserver
// service mounted at /preview (see main.go). Each segment is percent-encoded
// independently so the path separators stay intact.
export function previewUrl(path: string): string {
  const segments = path.split('/').filter(Boolean).map(encodeURIComponent)
  return '/preview/' + segments.join('/')
}
