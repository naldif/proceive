package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})
	}
}

// SuccessResponse merepresentasikan format JSON sukses standar
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// SuccessWithMetaResponse merepresentasikan format JSON sukses dengan metadata (pagination)
type SuccessWithMetaResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Data    interface{}    `json:"data"`
	Meta    PaginationMeta `json:"meta"`
}

// PaginationMeta merepresentasikan rincian paginasi data
type PaginationMeta struct {
	Page      int   `json:"page"`
	Limit     int   `json:"limit"`
	Total     int64 `json:"total"`
	TotalPage int   `json:"totalPage"`
	HasNext   bool  `json:"hasNext"`
	HasPrev   bool  `json:"hasPrev"`
}

// ErrorResponse merepresentasikan format JSON error standar dengan array error per field
type ErrorResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Errors  map[string][]string `json:"errors"`
}

// Success mengirimkan HTTP response sukses dengan pesan default "Berhasil"
func Success(c *gin.Context, httpCode int, data interface{}) {
	c.PureJSON(httpCode, SuccessResponse{
		Success: true,
		Message: "Berhasil",
		Data:    data,
	})
}

// SuccessWithMessage mengirimkan HTTP response sukses dengan pesan kustom dan data
func SuccessWithMessage(c *gin.Context, httpCode int, message string, data interface{}) {
	if message == "" {
		message = "Berhasil"
	}
	c.PureJSON(httpCode, SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SuccessWithMeta mengirimkan HTTP response sukses dengan pesan, data, dan metadata pagination
func SuccessWithMeta(c *gin.Context, httpCode int, message string, data interface{}, meta PaginationMeta) {
	if message == "" {
		message = "Berhasil mengambil data"
	}
	c.PureJSON(httpCode, SuccessWithMetaResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// SuccessWithPagination mengirimkan HTTP response sukses untuk daftar data berpaging
func SuccessWithPagination(c *gin.Context, httpCode int, message string, data interface{}, page, limit int, total int64) {
	if message == "" {
		message = "Berhasil mengambil data"
	}
	totalPages := 0
	if limit > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	meta := PaginationMeta{
		Page:      page,
		Limit:     limit,
		Total:     total,
		TotalPage: totalPages,
		HasNext:   page < totalPages,
		HasPrev:   page > 1,
	}

	SuccessWithMeta(c, httpCode, message, data, meta)
}

// Error mengirimkan HTTP response error umum / domain error
func Error(c *gin.Context, httpCode int, message string, errorsMap ...map[string][]string) {
	var errs map[string][]string
	if len(errorsMap) > 0 {
		errs = errorsMap[0]
	}

	c.PureJSON(httpCode, ErrorResponse{
		Success: false,
		Message: message,
		Errors:  errs,
	})
}

// ValidationError mengirimkan HTTP response error validasi dengan format array error per field
func ValidationError(c *gin.Context, message string, errorsMap map[string][]string) {
	if message == "" {
		message = "Validasi data gagal"
	}

	c.PureJSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Message: message,
		Errors:  errorsMap,
	})
}

// FormatValidationErrors mengonversi validator.ValidationErrors dari Gin atau error binding JSON menjadi map[string][]string
func FormatValidationErrors(err error) map[string][]string {
	errorsMap := make(map[string][]string)
	if err == nil {
		return errorsMap
	}

	var valErr validator.ValidationErrors
	if errors.As(err, &valErr) {
		for _, f := range valErr {
			field := f.Field()
			if field == "" {
				field = toSnakeCase(f.StructField())
			}
			msg := tagToMessage(field, f.Tag(), f.Param())
			errorsMap[field] = append(errorsMap[field], msg)
		}
		return errorsMap
	}

	// Tangani kasus type mismatch pada field tertentu saat unmarshal JSON
	var ute *json.UnmarshalTypeError
	if errors.As(err, &ute) {
		field := ute.Field
		if field == "" {
			field = extractFieldFromJSONError(err.Error())
		}
		if field == "" {
			field = "body"
		} else {
			field = toSnakeCase(field)
		}

		msg := fmt.Sprintf("Format atau tipe data %s tidak valid", field)
		if ute.Type != nil && strings.Contains(strings.ToLower(ute.Type.String()), "uuid") {
			msg = fmt.Sprintf("Format %s tidak valid (harus berupa UUID valid)", field)
		}
		errorsMap[field] = append(errorsMap[field], msg)
		return errorsMap
	}

	// Tangani kasus error unmarshal UUID atau tipe khusus yang tertulis di pesan error
	errStr := err.Error()
	if field := extractFieldFromJSONError(errStr); field != "" {
		field = toSnakeCase(field)
		msg := fmt.Sprintf("Format atau nilai %s tidak valid", field)
		if strings.Contains(strings.ToLower(errStr), "uuid") {
			msg = fmt.Sprintf("Format %s tidak valid (harus berupa UUID valid)", field)
		}
		errorsMap[field] = append(errorsMap[field], msg)
		return errorsMap
	}

	// Tangani kasus body request kosong (EOF)
	if errors.Is(err, io.EOF) {
		errorsMap["body"] = []string{"Body request tidak boleh kosong"}
		return errorsMap
	}

	// Tangani kasus sintaks JSON corrupt
	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		errorsMap["body"] = []string{"Format payload JSON tidak valid (sintaks JSON salah)"}
		return errorsMap
	}

	// Fallback umum dengan pesan error asli, gunakan key "body"
	errorsMap["body"] = []string{err.Error()}
	return errorsMap
}

var jsonFieldRegex = regexp.MustCompile(`(?i)(?:struct field\s+(?:\w+\.)?(\w+)|field\s+(\w+))`)

func extractFieldFromJSONError(errStr string) string {
	matches := jsonFieldRegex.FindStringSubmatch(errStr)
	if len(matches) > 1 && matches[1] != "" {
		return matches[1]
	}
	if len(matches) > 2 && matches[2] != "" {
		return matches[2]
	}
	return ""
}

func tagToMessage(field, tag, param string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("%s wajib diisi", field)
	case "email":
		return fmt.Sprintf("Format %s tidak valid", field)
	case "uuid", "uuid4":
		return fmt.Sprintf("Format %s tidak valid (harus berupa UUID)", field)
	case "min":
		return fmt.Sprintf("%s minimal %s karakter", field, param)
	case "max":
		return fmt.Sprintf("%s maksimal %s karakter", field, param)
	case "gte":
		return fmt.Sprintf("%s harus lebih besar atau sama dengan %s", field, param)
	case "lte":
		return fmt.Sprintf("%s harus lebih kecil atau sama dengan %s", field, param)
	case "gt":
		return fmt.Sprintf("%s harus lebih besar dari %s", field, param)
	case "lt":
		return fmt.Sprintf("%s harus lebih kecil dari %s", field, param)
	case "oneof":
		return fmt.Sprintf("%s harus berupa salah satu dari [%s]", field, param)
	default:
		return fmt.Sprintf("%s tidak memenuhi kriteria validasi (%s)", field, tag)
	}
}

func toSnakeCase(str string) string {
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	return strings.ToLower(matchAllCap.ReplaceAllString(snake, "${1}_${2}"))
}
