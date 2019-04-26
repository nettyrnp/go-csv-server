package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os"
)

const (
	idField = "N_REG_NEW"
)

var (
	logger, _     = zap.NewProduction()
	sugaredLogger = logger.Sugar()
)

type serviceResponse struct {
	Body  interface{}
	Error string
}

func main() {
	testParseMulti("АА3777РР", "testdata/tz_test.csv", "testdata/tz_test1.csv")
	return

	defer logger.Sync() // flushes buffer, if any

	conf, err := GetConfig()
	Die(err)
	fmt.Printf(">> Config: %v\n", conf)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	//r.Use(middleware.CloseNotify) // todo: uncomment
	r.Use(middleware.Timeout(conf.HTTP.ShutdownTimeout))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		enableFrontend(w, conf.HTTP.Frontend)
		sugaredLogger.Info(fmt.Sprintf("resp header: %v", w.Header()))
		w.Write([]byte("This is a root route"))
	})
	r.Get("/admin/version", func(w http.ResponseWriter, r *http.Request) {
		enableFrontend(w, conf.HTTP.Frontend)
		sugaredLogger.Info(fmt.Sprintf("resp header: %v", w.Header()))
		w.Write([]byte("0.0.0-00000"))
	})
	r.Get("/search", func(w http.ResponseWriter, r *http.Request) {
		svcResp := serviceResponse{}

		tname := r.URL.Query().Get("tname") // tname will be used in future
		if len(tname) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("param 'tname' is empty"))
			return
		}
		// Healthcheck
		if tname == "foo" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		sNumber := r.URL.Query().Get("snumber")
		if len(sNumber) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("param 'snumber' is empty"))
			return
		}
		enableFrontend(w, conf.HTTP.Frontend)

		oksources, err := searchBySourceNumber(sNumber)
		if err != nil {
			writeStatusBadRequest(w, svcResp, err.Error())
			return
		}
		fmt.Printf(">> Received a batch of %v OkSource rows\n", len(oksources))
		svcResp.Body = oksources

		host := r.Host
		userAgent := r.UserAgent()
		logTitle := "GET oksources by snumber"
		jsonResponse, _ := json.Marshal(svcResp)
		sugaredLogger.Info("search service", zap.Any(logTitle, map[string]interface{}{
			"hostFrom":  host,
			"userAgent": userAgent,
			"response":  svcResp,
		}))
		w.Write(jsonResponse)
	})

	addr := fmt.Sprintf("%s:%d", conf.HTTP.Host, conf.HTTP.Port)
	go func() {
		err := http.ListenAndServe(addr, r)
		Die(err)
	}()
	sugaredLogger.Infof("REST service is listening at URL: %s", addr)

	select {} // todo: graceful shutdown
}

func testParseMulti(id string, fnames ...string) {
	m2, err := toAggregatedMap(fnames)
	Die(err)
	fmt.Printf(">> m2.len: %v\n", len(m2))
	fmt.Printf(">> m2[%v].len: %v\n", id, len(m2[id]))
	for k, v := range m2[id] {
		fmt.Printf("\t>> %v:\t'%v'\t'%v'\n", k, v["REG_ADDR_KOATUU"], v["OPER_CODE"])
	}
}

func toAggregatedMap(fnames []string) (map[string][]map[string]string, error) {
	m2 := map[string][]map[string]string{}

	for _, fname := range fnames {
		m, err := toMap0(fname)
		if err != nil {
			return nil, err
		}
		for k, v := range m {
			v2, ok := m2[k]
			if !ok {
				m2[k] = []map[string]string{v}
			} else {
				v2 = append(v2, v)
				m2[k] = v2
			}
		}
	}

	return m2, nil
}

func toMap0(fname string) (map[string]map[string]string, error) {
	text, err := ReadFile(fname)
	//fmt.Printf(">> text: \n%v\n\n", text)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBufferString(text)
	arr := CSVToMap(buf)
	fmt.Printf(">> arr.len: %v\n", len(arr))

	m, err := toMap(arr)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func toMap(arr []map[string]string) (map[string]map[string]string, error) {
	m := map[string]map[string]string{}
	for _, row := range arr {
		id, ok := row[idField]
		if ok {
			m[id] = row
		} else {
			return nil, errors.Errorf("Now column '%v'", id)
		}
	}
	return m, nil
}

func CSVToMap(reader io.Reader) []map[string]string {
	r := csv.NewReader(reader)
	r.Comma = ';'
	rows := []map[string]string{}
	var header []string
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		Die(err)
		if header == nil {
			header = record
		} else {
			dict := map[string]string{}
			for i := range header {
				dict[header[i]] = record[i]
			}
			rows = append(rows, dict)
		}
	}
	return rows
}

func writeStatusBadRequest(w http.ResponseWriter, response serviceResponse, errorMsg string) {
	w.WriteHeader(http.StatusBadRequest)
	response.Error = errorMsg
	jsonResponse, _ := json.Marshal(response)
	w.Write(jsonResponse)
}

// CORS requires explicit naming of allowed URLs in response header
func enableFrontend(w http.ResponseWriter, url string) {
	w.Header().Set("Access-Control-Allow-Origin", url)
	w.Header().Set("Access-Control-Request-Method", "POST,GET,OPTIONS")
}

// Die kills the failing program.
func Die(err error) {
	if err == nil {
		return
	}
	//fmt.Println(">> "+err.Error())
	sugaredLogger.Fatal(err.Error())
	os.Exit(1)
}

func searchBySourceNumber(sNumber string) ([]string, error) {
	var arr []string
	return arr, nil
}
