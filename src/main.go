package main

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/tcnksm/go-httpstat"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type host struct {
	Addr string
	Code string
}

type data struct {
	Hosts []host
	Error string
}

var (
	dnsLookup = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "webapp",
			Name:      "dnsLookup",
			Help:      "Time spend to lookup DNS record",
		},
		[]string{"addr", "code"},
	)
	tcpConnection = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "webapp",
			Name:      "tcpConnection",
			Help:      "Time spend to connect",
		},
		[]string{"addr", "code"},
	)
	serverProcessing = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "webapp",
			Name:      "serverProcessing",
			Help:      "Time spend on wait of response",
		},
		[]string{"addr", "code"},
	)
	contentTransfer = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "webapp",
			Name:      "contentTransfer",
			Help:      "Time spend on waiting data from server",
		},
		[]string{"addr", "code"},
	)
	responseCodesFromHosts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "webapp",
			Name:      "responseCodesFromHosts",
			Help:      "Response codes from hosts",
		},
		[]string{"addr", "code"},
	)
)

func generateData() map[string]map[string]string {
	vars := map[string]map[string]string{
		"envs": make(map[string]string),
	}

	var envDirs string
	envDirs = os.Getenv("SECRETS_DIRS")
	if envDirs != "" {
		var dirs []string
		dirs = strings.Split(envDirs, ",")
		log.WithFields(log.Fields{
			"dirs": dirs,
		}).Info()

		var files []string

		for idx := range dirs {
			err := filepath.Walk(dirs[idx], func(path string, info os.FileInfo, err error) error {
				files = append(files, path)
				return nil
			})
			if err != nil {
				log.WithFields(log.Fields{
					"Error": err,
				}).Error()
			}
		}

		for idx := range files {
			pair := readVarFromFile(files[idx])

			if pair != nil {
				_, ok := vars["file"]
				if ok != true {
					vars["file"] = map[string]string{
						pair["key"]: pair["val"],
					}
				} else {
					vars["file"][pair["key"]] = pair["val"]
				}
			}
		}
	}

	for _, envs := range os.Environ() {
		pair := strings.SplitN(envs, "=", 2)
		vars["envs"][pair[0]] = pair[1]
	}

	return vars
}

func readVarFromFile(filename string) map[string]string {
	info, err := os.Stat(filename)
	if err != nil {
		log.Errorf("Something went wrong with access to file %v", filename)
		return nil
	}
	if info.IsDir() {
		return nil
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Errorf("Can't read from file %v", filename)
		return nil
	}

	pair := make(map[string]string)
	pair["key"] = strings.ToUpper(filepath.Base(filename))
	pair["val"] = string(data)
	return pair
}

func renderHtml(w http.ResponseWriter, r *http.Request) {
	vars := generateData()

	renderedPage, _ := template.ParseFiles("envs.gohtml")
	err := renderedPage.Execute(w, vars)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error()
	}

	log.WithFields(log.Fields{
		"Method":      r.Method,
		"URL":         r.URL.String(),
		"Remote-Addr": r.RemoteAddr,
	}).Info()
}

func ping(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status": "ok",
	}

	jsonData, marshallErr := json.Marshal(data)
	if marshallErr != nil {
		log.WithFields(log.Fields{
			"Error": marshallErr,
		}).Error()
	}
	w.Header().Set("Content-Type", "application/json")
	_, wErr := w.Write(jsonData)
	if wErr != nil {
		log.Errorf("An error occured: %s", wErr)
	}

	log.WithFields(log.Fields{
		"Method":      r.Method,
		"URL":         r.URL.String(),
		"Remote-Addr": r.RemoteAddr,
	}).Info()
}

func jsonEnvs(w http.ResponseWriter, r *http.Request) {
	data := generateData()

	jsonData, marshallErr := json.Marshal(data)
	if marshallErr != nil {
		log.WithFields(log.Fields{
			"Error": marshallErr,
		}).Error()
	}
	w.Header().Set("Content-Type", "application/json")
	_, wErr := w.Write(jsonData)
	if wErr != nil {
		log.Errorf("An error occured: %s", wErr)
	}

	log.WithFields(log.Fields{
		"Method":      r.Method,
		"URL":         r.URL.String(),
		"Remote-Addr": r.RemoteAddr,
	}).Info()
}

func httpQueryToHosts() data {
	var response data

	hostsList := os.Getenv("HTTP_HOSTS")
	if hostsList == "" {
		log.Warning("Variable HTTP_HOSTS is empty, should be list. Can not proceed http queries to hosts check")
		response.Error = "Variable HTTP_HOSTS is empty, should be list"
	} else {
		hosts := strings.Split(hostsList, ";")

		for _, hostAddr := range hosts {
			code := "0"
			req, respErr := http.NewRequest("GET", hostAddr, nil)
			if respErr != nil {
				log.WithFields(log.Fields{
					"Error": respErr,
				}).Error()
			}

			var result httpstat.Result
			ctx := httpstat.WithHTTPStat(req.Context(), &result)
			req = req.WithContext(ctx)
			// Send request by default HTTP client
			client := http.DefaultClient
			res, err := client.Do(req)
			if err != nil {
				log.Error(err)
			} else {
				code = strconv.Itoa(res.StatusCode)
				if _, err := io.Copy(ioutil.Discard, res.Body); err != nil {
					log.Error(err)
				}
				err := res.Body.Close()
				if err != nil {
					log.Errorf("Can not close response body %v", err)
				}
				end := time.Now()

				dnsLookup.With(prometheus.Labels{"addr": hostAddr, "code": code}).Observe(float64(result.DNSLookup / time.Millisecond))
				tcpConnection.With(prometheus.Labels{"addr": hostAddr, "code": code}).Observe(float64(result.TCPConnection / time.Millisecond))
				serverProcessing.With(prometheus.Labels{"addr": hostAddr, "code": code}).Observe(float64(result.ServerProcessing / time.Millisecond))
				contentTransfer.With(prometheus.Labels{"addr": hostAddr, "code": code}).Observe(float64(result.ContentTransfer(end) / time.Millisecond))
			}

			responseCodesFromHosts.With(prometheus.Labels{"addr": hostAddr, "code": code}).Inc()
			response.Hosts = append(response.Hosts, host{hostAddr, code})

			log.WithFields(log.Fields{
				"Host": hostAddr,
				"Code": code,
			}).Info()
		}
	}

	return response
}

func checkServices(w http.ResponseWriter, r *http.Request) {
	response := httpQueryToHosts

	jsonData, dataMarshallErr := json.Marshal(response)
	if dataMarshallErr != nil {
		log.WithFields(log.Fields{
			"Error": dataMarshallErr,
		}).Error()
	}

	w.Header().Set("Content-Type", "application/json")
	_, wErr := w.Write(jsonData)
	if wErr != nil {
		log.Errorf("An error occured: %s", wErr)
	}

	log.WithFields(log.Fields{
		"Method":      r.Method,
		"URL":         r.URL.String(),
		"Remote-Addr": r.RemoteAddr,
	}).Info()
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = "0.0.0.0:8080"
	}

	httpCheckPeriod := time.Duration(5) * time.Second
	httpCheckPeriodStr := os.Getenv("HTTP_CHECK_PERIOD")
	if httpCheckPeriodStr != "" {
		parsedInt, err := strconv.ParseInt(httpCheckPeriodStr, 0, 32)
		if err != nil {
			log.WithFields(log.Fields{
				"Error": err,
			}).Error()
		}
		httpCheckPeriod = time.Duration(parsedInt) * time.Second
	}
	log.Infof("Will check http hosts every %s", httpCheckPeriod)

	prometheus.MustRegister(dnsLookup)
	prometheus.MustRegister(tcpConnection)
	prometheus.MustRegister(serverProcessing)
	prometheus.MustRegister(contentTransfer)
	prometheus.MustRegister(responseCodesFromHosts)

	go func() {
		for true {
			log.Info("Start regular checks to HTTP hosts")
			httpQueryToHosts()
			time.Sleep(httpCheckPeriod)
		}
	}()

	log.Infof("HTTP server is starting on: %v", listenAddr)
	http.HandleFunc("/", renderHtml)
	http.HandleFunc("/json", jsonEnvs)
	http.HandleFunc("/ping", ping)
	http.HandleFunc("/net-check", checkServices)
	http.Handle("/metrics", promhttp.Handler())
	_ = http.ListenAndServe(listenAddr, nil)
}
