package proxy

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/runtime"
)

type accessLogEntry struct {
	Host      string
	Method    string
	Path      string
	Status    int
	RemoteIP  string
	UserAgent string
	CreatedAt time.Time
}

type requestLogger struct {
	pg     *runtime.PostgresClient
	ch     chan accessLogEntry
	stopCh chan struct{}
	wg     sync.WaitGroup
}

func newRequestLogger(pg *runtime.PostgresClient) *requestLogger {
	if pg == nil || pg.Pool == nil {
		return nil
	}
	l := &requestLogger{
		pg:     pg,
		ch:     make(chan accessLogEntry, 2048),
		stopCh: make(chan struct{}),
	}
	l.wg.Add(1)
	go l.run()
	return l
}

func (l *requestLogger) Close() {
	if l == nil {
		return
	}
	close(l.stopCh)
	l.wg.Wait()
}

func (l *requestLogger) Log(entry accessLogEntry) {
	if l == nil {
		return
	}
	select {
	case l.ch <- entry:
	default:
		log.Printf("request logger buffer full; dropping event host=%s path=%s", entry.Host, entry.Path)
	}
}

func (l *requestLogger) run() {
	defer l.wg.Done()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	batch := make([]accessLogEntry, 0, 250)
	flush := func(ctx context.Context) {
		if len(batch) == 0 {
			return
		}
		if err := l.insertBatch(ctx, batch); err != nil {
			log.Printf("request log batch insert failed: %v", err)
		}
		batch = batch[:0]
	}

	for {
		select {
		case <-l.stopCh:
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			flush(ctx)
			cancel()
			return
		case rec := <-l.ch:
			batch = append(batch, rec)
			if len(batch) >= 250 {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				flush(ctx)
				cancel()
			}
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			flush(ctx)
			cancel()
		}
	}
}

func (l *requestLogger) insertBatch(ctx context.Context, batch []accessLogEntry) error {
	values := make([]string, 0, len(batch))
	args := make([]any, 0, len(batch)*7)
	for i, entry := range batch {
		offset := i*7 + 1
		values = append(values, fmt.Sprintf("($%d::uuid, $%d, $%d, $%d, $%d, $%d, $%d)", offset, offset+1, offset+2, offset+3, offset+4, offset+5, offset+6))
		args = append(args,
			newUUID(),
			normalizeDBText(entry.Host),
			normalizeDBText(entry.Method),
			normalizeDBText(entry.Path),
			entry.Status,
			normalizeDBText(entry.RemoteIP),
			normalizeDBText(entry.UserAgent),
		)
	}
	_, err := l.pg.Pool.Exec(
		ctx,
		`INSERT INTO request_logs (id, host, method, path, status_code, remote_ip, user_agent) VALUES `+strings.Join(values, ","),
		args...,
	)
	return err
}

func normalizeDBText(v string) string {
	return strings.TrimSpace(v)
}

func newUUID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	buf := make([]byte, 36)
	hex.Encode(buf[0:8], b[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], b[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], b[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], b[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:36], b[10:16])
	return string(buf)
}

type statusCaptureWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusCaptureWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusCaptureWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}
