package libreoffice

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/gotenberg/gotenberg/v7/pkg/gotenberg"
	"github.com/gotenberg/gotenberg/v7/pkg/modules/api"
	libreofficeapi "github.com/gotenberg/gotenberg/v7/pkg/modules/libreoffice/api"
)

func TestConvertRoute(t *testing.T) {
	for _, tc := range []struct {
		scenario               string
		ctx                    *api.ContextMock
		libreOffice            libreofficeapi.Uno
		engine                 gotenberg.PDFEngine
		expectOptions          libreofficeapi.Options
		expectError            bool
		expectHttpError        bool
		expectHttpStatus       int
		expectOutputPathsCount int
	}{
		{
			scenario: "missing at least one mandatory file",
			ctx:      &api.ContextMock{Context: new(api.Context)},
			libreOffice: &libreofficeapi.ApiMock{ExtensionsMock: func() []string {
				return []string{".docx"}
			}},
			expectError:            true,
			expectHttpError:        true,
			expectHttpStatus:       http.StatusBadRequest,
			expectOutputPathsCount: 0,
		},
		{
			scenario: "ErrMalformedPageRanges",
			ctx: func() *api.ContextMock {
				ctx := &api.ContextMock{Context: new(api.Context)}
				ctx.SetFiles(map[string]string{
					"document.docx": "/document.docx",
				})
				return ctx
			}(),
			libreOffice: &libreofficeapi.ApiMock{
				PdfMock: func(ctx context.Context, logger *zap.Logger, inputPath, outputPath string, options libreofficeapi.Options) error {
					return libreofficeapi.ErrMalformedPageRanges
				},
				ExtensionsMock: func() []string {
					return []string{".docx"}
				},
			},
			expectError:            true,
			expectHttpError:        true,
			expectHttpStatus:       http.StatusBadRequest,
			expectOutputPathsCount: 0,
		},
		{
			scenario: "error from LibreOffice",
			ctx: func() *api.ContextMock {
				ctx := &api.ContextMock{Context: new(api.Context)}
				ctx.SetFiles(map[string]string{
					"document.docx": "/document.docx",
				})
				return ctx
			}(),
			libreOffice: &libreofficeapi.ApiMock{
				PdfMock: func(ctx context.Context, logger *zap.Logger, inputPath, outputPath string, options libreofficeapi.Options) error {
					return errors.New("foo")
				},
				ExtensionsMock: func() []string {
					return []string{".docx"}
				},
			},
			expectError:            true,
			expectHttpError:        false,
			expectOutputPathsCount: 0,
		},
		{
			scenario: "ErrPDFFormatNotAvailable (single file)",
			ctx: func() *api.ContextMock {
				ctx := &api.ContextMock{Context: new(api.Context)}
				ctx.SetFiles(map[string]string{
					"document.docx": "/document.docx",
				})
				ctx.SetValues(map[string][]string{
					"pdfFormat": {
						"foo",
					},
				})
				return ctx
			}(),
			libreOffice: &libreofficeapi.ApiMock{
				PdfMock: func(ctx context.Context, logger *zap.Logger, inputPath, outputPath string, options libreofficeapi.Options) error {
					return nil
				},
				ExtensionsMock: func() []string {
					return []string{".docx"}
				},
			},
			engine: &gotenberg.PDFEngineMock{
				ConvertMock: func(ctx context.Context, logger *zap.Logger, format, inputPath, outputPath string) error {
					return gotenberg.ErrPDFFormatNotAvailable
				},
			},
			expectError:            true,
			expectHttpError:        true,
			expectHttpStatus:       http.StatusBadRequest,
			expectOutputPathsCount: 0,
		},
		{
			scenario: "PDF engine convert error (single file)",
			ctx: func() *api.ContextMock {
				ctx := &api.ContextMock{Context: new(api.Context)}
				ctx.SetFiles(map[string]string{
					"document.docx": "/document.docx",
				})
				ctx.SetValues(map[string][]string{
					"pdfFormat": {
						gotenberg.FormatPDFA1a,
					},
				})
				return ctx
			}(),
			libreOffice: &libreofficeapi.ApiMock{
				PdfMock: func(ctx context.Context, logger *zap.Logger, inputPath, outputPath string, options libreofficeapi.Options) error {
					return nil
				},
				ExtensionsMock: func() []string {
					return []string{".docx"}
				},
			},
			engine: &gotenberg.PDFEngineMock{
				ConvertMock: func(ctx context.Context, logger *zap.Logger, format, inputPath, outputPath string) error {
					return errors.New("foo")
				},
			},
			expectError:            true,
			expectHttpError:        false,
			expectOutputPathsCount: 0,
		},
		{
			scenario: "cannot add output paths (single file)",
			ctx: func() *api.ContextMock {
				ctx := &api.ContextMock{Context: new(api.Context)}
				ctx.SetFiles(map[string]string{
					"document.docx": "/document.docx",
				})
				ctx.SetCancelled(true)
				return ctx
			}(),
			libreOffice: &libreofficeapi.ApiMock{
				PdfMock: func(ctx context.Context, logger *zap.Logger, inputPath, outputPath string, options libreofficeapi.Options) error {
					return nil
				},
				ExtensionsMock: func() []string {
					return []string{".docx"}
				},
			},
			expectError:            true,
			expectHttpError:        false,
			expectOutputPathsCount: 0,
		},
		{
			scenario: "success (single file)",
			ctx: func() *api.ContextMock {
				ctx := &api.ContextMock{Context: new(api.Context)}
				ctx.SetFiles(map[string]string{
					"document.docx": "/document.docx",
				})
				return ctx
			}(),
			libreOffice: &libreofficeapi.ApiMock{
				PdfMock: func(ctx context.Context, logger *zap.Logger, inputPath, outputPath string, options libreofficeapi.Options) error {
					return nil
				},
				ExtensionsMock: func() []string {
					return []string{".docx"}
				},
			},
			expectError:            false,
			expectHttpError:        false,
			expectOutputPathsCount: 1,
		},
		{
			scenario: "success (many files)",
			ctx: func() *api.ContextMock {
				ctx := &api.ContextMock{Context: new(api.Context)}
				ctx.SetFiles(map[string]string{
					"document.docx":  "/document.docx",
					"document2.docx": "/document2.docx",
				})
				return ctx
			}(),
			libreOffice: &libreofficeapi.ApiMock{
				PdfMock: func(ctx context.Context, logger *zap.Logger, inputPath, outputPath string, options libreofficeapi.Options) error {
					return nil
				},
				ExtensionsMock: func() []string {
					return []string{".docx"}
				},
			},
			expectError:            false,
			expectHttpError:        false,
			expectOutputPathsCount: 2,
		},
		{
			scenario: "success with PDF format (single file)",
			ctx: func() *api.ContextMock {
				ctx := &api.ContextMock{Context: new(api.Context)}
				ctx.SetFiles(map[string]string{
					"document.docx": "/document.docx",
				})
				ctx.SetValues(map[string][]string{
					"pdfFormat": {
						gotenberg.FormatPDFA1a,
					},
				})
				return ctx
			}(),
			libreOffice: &libreofficeapi.ApiMock{
				PdfMock: func(ctx context.Context, logger *zap.Logger, inputPath, outputPath string, options libreofficeapi.Options) error {
					return nil
				},
				ExtensionsMock: func() []string {
					return []string{".docx"}
				},
			},
			engine: &gotenberg.PDFEngineMock{
				ConvertMock: func(ctx context.Context, logger *zap.Logger, format, inputPath, outputPath string) error {
					return nil
				},
			},
			expectError:            false,
			expectHttpError:        false,
			expectOutputPathsCount: 1,
		},
		{
			scenario: "success with every PDF formats form field (single file)",
			ctx: func() *api.ContextMock {
				ctx := &api.ContextMock{Context: new(api.Context)}
				ctx.SetFiles(map[string]string{
					"document.docx": "/document.docx",
				})
				ctx.SetValues(map[string][]string{
					"nativePdfA1aFormat": {
						"true",
					},
					"pdfFormat": {
						gotenberg.FormatPDFA1a,
					},
					"nativePdfFormat": {
						gotenberg.FormatPDFA1a,
					},
				})
				return ctx
			}(),
			libreOffice: &libreofficeapi.ApiMock{
				PdfMock: func(ctx context.Context, logger *zap.Logger, inputPath, outputPath string, options libreofficeapi.Options) error {
					return nil
				},
				ExtensionsMock: func() []string {
					return []string{".docx"}
				},
			},
			engine: &gotenberg.PDFEngineMock{
				ConvertMock: func(ctx context.Context, logger *zap.Logger, format, inputPath, outputPath string) error {
					return nil
				},
			},
			expectError:            false,
			expectHttpError:        false,
			expectOutputPathsCount: 1,
		},
		{
			scenario: "merge error",
			ctx: func() *api.ContextMock {
				ctx := &api.ContextMock{Context: new(api.Context)}
				ctx.SetFiles(map[string]string{
					"document.docx":  "/document.docx",
					"document2.docx": "/document2.docx",
				})
				ctx.SetValues(map[string][]string{
					"merge": {
						"true",
					},
				})
				return ctx
			}(),
			libreOffice: &libreofficeapi.ApiMock{
				PdfMock: func(ctx context.Context, logger *zap.Logger, inputPath, outputPath string, options libreofficeapi.Options) error {
					return nil
				},
				ExtensionsMock: func() []string {
					return []string{".docx"}
				},
			},
			engine: &gotenberg.PDFEngineMock{
				MergeMock: func(ctx context.Context, logger *zap.Logger, inputPaths []string, outputPath string) error {
					return errors.New("foo")
				},
			},
			expectError:            true,
			expectHttpError:        false,
			expectOutputPathsCount: 0,
		},
		{
			scenario: "ErrPDFFormatNotAvailable (merge)",
			ctx: func() *api.ContextMock {
				ctx := &api.ContextMock{Context: new(api.Context)}
				ctx.SetFiles(map[string]string{
					"document.docx":  "/document.docx",
					"document2.docx": "/document2.docx",
				})
				ctx.SetValues(map[string][]string{
					"merge": {
						"true",
					},
					"pdfFormat": {
						"foo",
					},
				})
				return ctx
			}(),
			libreOffice: &libreofficeapi.ApiMock{
				PdfMock: func(ctx context.Context, logger *zap.Logger, inputPath, outputPath string, options libreofficeapi.Options) error {
					return nil
				},
				ExtensionsMock: func() []string {
					return []string{".docx"}
				},
			},
			engine: &gotenberg.PDFEngineMock{
				MergeMock: func(ctx context.Context, logger *zap.Logger, inputPaths []string, outputPath string) error {
					return nil
				},
				ConvertMock: func(ctx context.Context, logger *zap.Logger, format, inputPath, outputPath string) error {
					return gotenberg.ErrPDFFormatNotAvailable
				},
			},
			expectError:            true,
			expectHttpError:        true,
			expectHttpStatus:       http.StatusBadRequest,
			expectOutputPathsCount: 0,
		},
		{
			scenario: "PDF engine convert error (merge)",
			ctx: func() *api.ContextMock {
				ctx := &api.ContextMock{Context: new(api.Context)}
				ctx.SetFiles(map[string]string{
					"document.docx":  "/document.docx",
					"document2.docx": "/document2.docx",
				})
				ctx.SetValues(map[string][]string{
					"merge": {
						"true",
					},
					"pdfFormat": {
						gotenberg.FormatPDFA1a,
					},
				})
				return ctx
			}(),
			libreOffice: &libreofficeapi.ApiMock{
				PdfMock: func(ctx context.Context, logger *zap.Logger, inputPath, outputPath string, options libreofficeapi.Options) error {
					return nil
				},
				ExtensionsMock: func() []string {
					return []string{".docx"}
				},
			},
			engine: &gotenberg.PDFEngineMock{
				MergeMock: func(ctx context.Context, logger *zap.Logger, inputPaths []string, outputPath string) error {
					return nil
				},
				ConvertMock: func(ctx context.Context, logger *zap.Logger, format, inputPath, outputPath string) error {
					return errors.New("foo")
				},
			},
			expectError:            true,
			expectHttpError:        false,
			expectOutputPathsCount: 0,
		},
		{
			scenario: "cannot add output paths (merge)",
			ctx: func() *api.ContextMock {
				ctx := &api.ContextMock{Context: new(api.Context)}
				ctx.SetFiles(map[string]string{
					"document.docx":  "/document.docx",
					"document2.docx": "/document2.docx",
				})
				ctx.SetValues(map[string][]string{
					"merge": {
						"true",
					},
				})
				ctx.SetCancelled(true)
				return ctx
			}(),
			libreOffice: &libreofficeapi.ApiMock{
				PdfMock: func(ctx context.Context, logger *zap.Logger, inputPath, outputPath string, options libreofficeapi.Options) error {
					return nil
				},
				ExtensionsMock: func() []string {
					return []string{".docx"}
				},
			},
			engine: &gotenberg.PDFEngineMock{
				MergeMock: func(ctx context.Context, logger *zap.Logger, inputPaths []string, outputPath string) error {
					return nil
				},
			},
			expectError:            true,
			expectHttpError:        false,
			expectOutputPathsCount: 0,
		},
		{
			scenario: "success (merge)",
			ctx: func() *api.ContextMock {
				ctx := &api.ContextMock{Context: new(api.Context)}
				ctx.SetFiles(map[string]string{
					"document.docx":  "/document.docx",
					"document2.docx": "/document2.docx",
				})
				ctx.SetValues(map[string][]string{
					"merge": {
						"true",
					},
				})
				return ctx
			}(),
			libreOffice: &libreofficeapi.ApiMock{
				PdfMock: func(ctx context.Context, logger *zap.Logger, inputPath, outputPath string, options libreofficeapi.Options) error {
					return nil
				},
				ExtensionsMock: func() []string {
					return []string{".docx"}
				},
			},
			engine: &gotenberg.PDFEngineMock{
				MergeMock: func(ctx context.Context, logger *zap.Logger, inputPaths []string, outputPath string) error {
					return nil
				},
			},
			expectError:            false,
			expectHttpError:        false,
			expectOutputPathsCount: 1,
		},
		{
			scenario: "success with PDF format (merge)",
			ctx: func() *api.ContextMock {
				ctx := &api.ContextMock{Context: new(api.Context)}
				ctx.SetFiles(map[string]string{
					"document.docx":  "/document.docx",
					"document2.docx": "/document2.docx",
				})
				ctx.SetValues(map[string][]string{
					"merge": {
						"true",
					},
					"pdfFormat": {
						gotenberg.FormatPDFA1a,
					},
				})
				return ctx
			}(),
			libreOffice: &libreofficeapi.ApiMock{
				PdfMock: func(ctx context.Context, logger *zap.Logger, inputPath, outputPath string, options libreofficeapi.Options) error {
					return nil
				},
				ExtensionsMock: func() []string {
					return []string{".docx"}
				},
			},
			engine: &gotenberg.PDFEngineMock{
				MergeMock: func(ctx context.Context, logger *zap.Logger, inputPaths []string, outputPath string) error {
					return nil
				},
				ConvertMock: func(ctx context.Context, logger *zap.Logger, format, inputPath, outputPath string) error {
					return nil
				},
			},
			expectError:            false,
			expectHttpError:        false,
			expectOutputPathsCount: 1,
		},
	} {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.ctx.SetLogger(zap.NewNop())
			c := echo.New().NewContext(nil, nil)
			c.Set("context", tc.ctx.Context)

			err := convertRoute(tc.libreOffice, tc.engine).Handler(c)

			if tc.expectError && err == nil {
				t.Fatal("expected error but got none", err)
			}

			if !tc.expectError && err != nil {
				t.Fatalf("expected no error but got: %v", err)
			}

			var httpErr api.HTTPError
			isHTTPErr := errors.As(err, &httpErr)

			if tc.expectHttpError && !isHTTPErr {
				t.Errorf("expected an HTTP error but got: %v", err)
			}

			if !tc.expectHttpError && isHTTPErr {
				t.Errorf("expected no HTTP error but got one: %v", httpErr)
			}

			if err != nil && tc.expectHttpError && isHTTPErr {
				status, _ := httpErr.HTTPError()
				if status != tc.expectHttpStatus {
					t.Errorf("expected %d as HTTP status code but got %d", tc.expectHttpStatus, status)
				}
			}

			if tc.expectOutputPathsCount != len(tc.ctx.OutputPaths()) {
				t.Errorf("expected %d output paths but got %d", tc.expectOutputPathsCount, len(tc.ctx.OutputPaths()))
			}
		})
	}
}
