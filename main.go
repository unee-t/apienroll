package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	jsonhandler "github.com/apex/log/handlers/json"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tj/go/http/response"
	
	"github.com/unee-t/env"

	"github.com/apex/log"
	"github.com/apex/log/handlers/text"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

var pingPollingFreq = 5 * time.Second

type handler struct {
	DSN            string
	APIAccessToken string
	db             *sql.DB
	Code           env.EnvCode
}

// APIKey is defined by the Table: user_api_keys https://s.natalian.org/2018-06-01/1527810246_2558x1406.png
type APIkey struct {
	UserID     string `json:"UserId"`
	UserAPIkey string `json:"userApiKey"`
}

func init() {
	if os.Getenv("UP_STAGE") == "" {
		log.SetHandler(text.Default)
	} else {
		log.SetHandler(jsonhandler.Default)
	}
}

// NewDbConnexion setups the configuration assuming various parameters have been setup in the AWS account
// TODO: REPLACE WITH THE `env.NewBzDbConnexion` FUNCTION
func NewDbConnexion() (h handler, err error) {

	// We check if the AWS CLI profile we need has been setup in this environment
		awsCliProfile, ok := os.LookupEnv("TRAVIS_AWS_PROFILE")
		if ok {
			log.Infof("NewDbConnexion Log: the AWS CLI profile we use is: %s", awsCliProfile)
		} else {
			log.Fatal("NewDbConnexion Fatal: the AWS CLI profile is unset as an environment variable, this is a fatal problem")
		}

		cfg, err := external.LoadDefaultAWSConfig(external.WithSharedConfigProfile(awsCliProfile))
		if err != nil {
			log.WithError(err).Fatal("NewDbConnexion Fatal: We do not have the AWS credentials we need")
			return
		}

	// We get the value for the DEFAULT_REGION
		defaultRegion, ok := os.LookupEnv("DEFAULT_REGION")
		if ok {
			log.Infof("NewDbConnexion Log: DEFAULT_REGION was overridden by local env: %s", defaultRegion)
		} else {
			log.Fatal("NewDbConnexion Fatal: DEFAULT_REGION is unset as an environment variable, this is a fatal problem")
		}

		cfg.Region = defaultRegion
		log.Infof("NewDbConnexion Log: The AWS region for this environment has been set to: %s", cfg.Region)

	// We get the value for the API_ACCESS_TOKEN
		apiAccessToken, ok := os.LookupEnv("API_ACCESS_TOKEN")
		if ok {
			log.Infof("NewDbConnexion Log: API_ACCESS_TOKEN was overridden by local env: **hidden secret**")
		} else {
			log.Fatal("NewDbConnexion Fatal: API_ACCESS_TOKEN is unset as an environment variable, this is a fatal problem")
		}

	e, err := env.NewConfig(cfg)
	if err != nil {
		log.WithError(err).Warn("NewDbConnexion Warning: error getting some of the parameters for that environment")
	}

	h = handler{
		DSN:            e.BugzillaDSN(), // `BugzillaDSN` is a function that is defined in the uneet/env/main.go dependency.
		APIAccessToken: apiAccessToken,
		Code:           e.Code,
	}

	h.db, err = sql.Open("mysql", h.DSN)
	if err != nil {
		log.WithError(err).Fatal("NewDbConnexion fatal: error opening database")
		return
	}

	microservicecheck := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "microservice",
			Help: "Version with DB ping check",
		},
		[]string{
			"commit",
		},
	)

	version := os.Getenv("UP_COMMIT")

	go func() {
		for {
			if h.db.Ping() == nil {
				microservicecheck.WithLabelValues(version).Set(1)
			} else {
				microservicecheck.WithLabelValues(version).Set(0)
			}
			time.Sleep(pingPollingFreq)
		}
	}()

	err = prometheus.Register(microservicecheck)
	if err != nil {
		log.Warn("NewDbConnexion Warning: prom already registered")
	}
	return
}

func main() {

	h, err := NewDbConnexion()
	if err != nil {
		log.WithError(err).Fatal("main Error: We are not able to connect to the BZ database")
		return
	}

	defer h.db.Close()

	addr := ":" + os.Getenv("PORT")

	app := mux.NewRouter()
	app.HandleFunc("/", h.enroll).Methods("POST")
	app.HandleFunc("/", h.ping).Methods("GET")

	if err := http.ListenAndServe(addr, env.Protect(app, h.APIAccessToken)); err != nil {
		log.WithError(err).Fatal("main Error: We have an error listening to http - API token has been set")
	}

}

func (h handler) insert(credential APIkey) (err error) {
	_, err = h.db.Exec(
		`INSERT INTO user_api_keys (user_id,
			api_key,
			description
		) VALUES (?,?,?)`,
		credential.UserID,
		credential.UserAPIkey,
		"MEFE Access Key",
	)
	return
}

func (h handler) enroll(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	var k APIkey
	err := decoder.Decode(&k)

	if err != nil {
		log.WithError(err).Errorf("enroll Error: We have an Input error - JSON is invalid")
		response.BadRequest(w, "enroll BadRequest: The request uses Invalid JSON")
		return
	}
	defer r.Body.Close()

	ctx := log.WithFields(log.Fields{
		"APIkey": k,
	})

	ctx.Info("Decoded (whatever this means...)")

	if k.UserAPIkey == "" {
		response.BadRequest(w, "enroll BadRequest: We are missing the APIkey that we need to insert")
		return
	}

	if k.UserID == "" {
		response.BadRequest(w, "enroll BadRequest: We are missing the BZ UserID")
		return
	}

	err = h.insert(k)

	if err != nil {
		log.WithError(err).Warnf("enroll Warning: We were not able to insert the API key for the new user in the BZ database")
		response.BadRequest(w, "enroll BadRequest: We were not able to insert the API key for the new user in the BZ database")
		return
	}

	response.OK(w)
	return

}

func (h handler) ping(w http.ResponseWriter, r *http.Request) {
	err := h.db.Ping()
	if err != nil {
		log.WithError(err).Error("ping Error: we have not been able to ping the BZ database")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	fmt.Fprintf(w, "OK - we are able to ping the BZ database")
}
