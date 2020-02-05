package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	jsonhandler "github.com/apex/log/handlers/json"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tj/go/http/response"
	
	//"github.com/unee-t/env"//

	"github.com/apex/log"
	"github.com/apex/log/handlers/text"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	// DEBUGGING
	"context"
	"strings"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	// END DEBUGGING
)

var pingPollingFreq = 5 * time.Second

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

// DEBUGGING - Move the code that belongs to unee-t/env here to facilitate debugging

type EnvCode int

type handler struct {
	DSN            string // aurora database connection string
	APIAccessToken string
	db             *sql.DB
	Code           EnvCode
}

// environment is the data type to manage the different environment (or STAGE) for a given Unee-T installation
type environment struct {
	Code      EnvCode
	Cfg       aws.Config
	AccountID string
	Stage     string
}

// https://github.com/unee-t/processInvitations/blob/master/sql/1_process_one_invitation_all_scenario_v3.0.sql#L12-L16
const (
	EnvUnknown EnvCode = iota // Oops
	EnvDev                    // Development aka Staging
	EnvProd                   // Production
	EnvDemo                   // Demo, which is like Production, for prospective customers to try
)

// GetSecret is the Golang equivalent for
// aws --profile your-aws-cli-profile ssm get-parameters --names API_ACCESS_TOKEN --with-decryption --query Parameters[0].Value --output text

func (e environment) GetSecret(key string) string {

	val, ok := os.LookupEnv(key)
	if ok {
		log.Warnf("GetSecret Warning: No need to query AWS parameter store: %s overridden by local env", key)
		return val
	}
	// Ideally environment above is set to avoid costly ssm (parameter store) lookups

	ps := ssm.New(e.Cfg)
	in := &ssm.GetParameterInput{
		Name:           aws.String(key),
		WithDecryption: aws.Bool(true),
	}
	req := ps.GetParameterRequest(in)
	out, err := req.Send(context.TODO())
	if err != nil {
		log.WithError(err).Errorf("GetSecret Error: failed to retrieve credentials for looking up %s", key)
		return ""
	}
	return aws.StringValue(out.Parameter.Value)
}

// NewConfig setups the configuration assuming various parameters have been setup in the AWS account
// - DEFAULT_REGION
// - STAGE
func NewConfig(cfg aws.Config) (e environment, err error) {

	// Save for ssm
		e.Cfg = cfg

		svc := sts.New(cfg)
		input := &sts.GetCallerIdentityInput{}
		req := svc.GetCallerIdentityRequest(input)
		result, err := req.Send(context.TODO())
		if err != nil {
			return e, err
		}

	// We get the ID of the AWS account we use
		e.AccountID = aws.StringValue(result.Account)
		log.Infof("NewConfig Log: The AWS Account ID for this environment is: %s", e.AccountID)

	// We get the value for the DEFAULT_REGION
		var defaultRegion string
		valdefaultRegion, ok := os.LookupEnv("DEFAULT_REGION")
		if ok {
			defaultRegion = valdefaultRegion
			log.Infof("NewConfig Log: DEFAULT_REGION was overridden by local env: %s", valdefaultRegion)
		} else {
			defaultRegion = e.GetSecret("DEFAULT_REGION")
			log.Infof("NewConfig Log: We get the DEFAULT_REGION from the AWS parameter store")
		}
	
		if defaultRegion == "" {
			log.Fatal("NewConfig fatal: DEFAULT_REGION is unset, this is a fatal problem")
		}

		cfg.Region = defaultRegion
		log.Infof("NewConfig Log: The AWS region for this environment has been set to: %s", cfg.Region)

	// We get the value for the STAGE
		var stage string
		valstage, ok := os.LookupEnv("STAGE")
		if ok {
			stage = valstage
			log.Infof("NewConfig Log: STAGE was overridden by local env: %s", valstage)
		} else {
			defaultRegion = e.GetSecret("STAGE")
			log.Infof("NewConfig Log:  We get the STAGE from the AWS parameter store")
		}
	
		if stage == "" {
			log.Fatal("NewConfig fatal: STAGE is unset, this is a fatal problem")
		}

		e.Stage = stage

	// Based on the value of the STAGE variable we do different things
		switch e.Stage {
		case "dev":
			e.Code = EnvDev
			return e, nil
		case "prod":
			e.Code = EnvProd
			return e, nil
		case "demo":
			e.Code = EnvDemo
			return e, nil
		default:
			log.WithField("stage", e.Stage).Error("NewConfig Error: unknown stage")
			return e, nil
		}
}

func (e environment) BugzillaDSN() string {

	// Get the value of the variable BUGZILLA_DB_USER
		var bugzillaDbUser string
		valbugzillaDbUser, ok := os.LookupEnv("BUGZILLA_DB_USER")
		if ok {
			bugzillaDbUser = valbugzillaDbUser
			log.Infof("BugzillaDSN Log: BUGZILLA_DB_USER was overridden by local env: %s", valbugzillaDbUser)
		} else {
			bugzillaDbUser = e.GetSecret("BUGZILLA_DB_USER")
			log.Infof("BugzillaDSN Log: We get the BUGZILLA_DB_USER from the AWS parameter store")
		}

		if bugzillaDbUser == "" {
			log.Fatal("BugzillaDSN Fatal: BUGZILLA_DB_USER is unset, this is a fatal problem")
		}

	// Get the value of the variable 
		var bugzillaDbPassword string
		valbugzillaDbPassword, ok := os.LookupEnv("BUGZILLA_DB_PASSWORD")
		if ok {
			bugzillaDbPassword = valbugzillaDbPassword
			log.Infof("BugzillaDSN Log: BUGZILLA_DB_PASSWORD was overridden by local env: **hidden_secret**")
		} else {
			bugzillaDbPassword = e.GetSecret("BUGZILLA_DB_PASSWORD")
			log.Infof("BugzillaDSN Log: We get the BUGZILLA_DB_PASSWORD from the AWS parameter store")
		}

		if bugzillaDbPassword == "" {
			log.Fatal("BugzillaDSN Fatal: BUGZILLA_DB_PASSWORD is unset, this is a fatal problem")
		}

	// Get the value of the variable 
		var mysqlhost string
		valmysqlhost, ok := os.LookupEnv("MYSQL_HOST")
		if ok {
			mysqlhost = valmysqlhost
			log.Infof("BugzillaDSN Log: MYSQL_HOST was overridden by local env: %s", valmysqlhost)
		} else {
			mysqlhost = e.GetSecret("MYSQL_HOST")
			log.Infof("BugzillaDSN Log: We get the MYSQL_HOST from the AWS parameter store")
		}

		if mysqlhost == "" {
			log.Fatal("BugzillaDSN Fatal: MYSQL_HOST is unset, this is a fatal problem")
		}

	// Get the value of the variable 
		var mysqlport string
		valmysqlport, ok := os.LookupEnv("MYSQL_PORT")
		if ok {
			mysqlport = valmysqlport
			log.Infof("BugzillaDSN Log: MYSQL_PORT was overridden by local env: %s", valmysqlport)
		} else {
			mysqlport = e.GetSecret("MYSQL_PORT")
			log.Infof("BugzillaDSN Log: We get the MYSQL_PORT from the AWS parameter store")
		}

		if mysqlport == "" {
			log.Fatal("BugzillaDSN Fatal: MYSQL_PORT is unset, this is a fatal problem")
		}

	// Get the value of the variable 
		var bugzillaDbName string
		valbugzillaDbName, ok := os.LookupEnv("BUGZILLA_DB_NAME")
		if ok {
			bugzillaDbName = valbugzillaDbName
			log.Infof("BugzillaDSN Log: BUGZILLA_DB_NAME was overridden by local env: %s", valbugzillaDbName)
		} else {
			bugzillaDbName = e.GetSecret("BUGZILLA_DB_NAME")
			log.Infof("BugzillaDSN Log: We get the BUGZILLA_DB_NAME from the AWS parameter store")
		}

		if bugzillaDbName == "" {
			log.Fatal("BugzillaDSN Fatal: BUGZILLA_DB_NAME is unset, this is a fatal problem")
		}

	// Build the string that will allow connection to the BZ database
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?multiStatements=true&sql_mode=TRADITIONAL&timeout=15s&collation=utf8mb4_unicode_520_ci",
			bugzillaDbUser,
			bugzillaDbPassword,
			mysqlhost,
			mysqlport,
			bugzillaDbName)
}

// Protect using: curl -H 'Authorization: Bearer secret' style
// Modelled after https://github.com/apex/up-examples/blob/master/oss/golang-basic-auth/main.go#L16
func Protect(h http.Handler, APIAccessToken string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var token string
		// Get token from the Authorization header
		// format: Authorization: Bearer
		tokens, ok := r.Header["Authorization"]
		if ok && len(tokens) >= 1 {
			token = tokens[0]
			token = strings.TrimPrefix(token, "Bearer ")
		}
		if token == "" || token != APIAccessToken {
			log.Errorf("Protect Error: Token %q != APIAccessToken %q", token, APIAccessToken)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// Towr is a workaround for gorilla/pat: https://stackoverflow.com/questions/50753049/
// Wish I could make this simpler
func Towr(h http.Handler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) { h.ServeHTTP(w, r) }
}


// END DEBUGGING



// NewDbConnexion setups the configuration assuming various parameters have been setup in the AWS account
// TODO: REPLACE WITH THE `env.NewBzDbConnexion` FUNCTION
func NewDbConnexion() (h handler, err error) {

	// We get the AWS configuration information for the default profile

		cfg, err := external.LoadDefaultAWSConfig()
		if err != nil {
			log.WithError(err).Fatal("NewDbConnexion Fatal: We do not have the AWS credentials we need")
			return
		}

		// cfg also needs the default region.
		// We get the value for the DEFAULT_REGION
			defaultRegion, ok := os.LookupEnv("DEFAULT_REGION")
			if ok {
				log.Infof("NewDbConnexion Log: DEFAULT_REGION was overridden by local env: %s", defaultRegion)
			} else {
				log.Fatal("NewDbConnexion Fatal: DEFAULT_REGION is unset as an environment variable, this is a fatal problem")
			}
			
			// Set the AWS Region that the service clients should use
			cfg.Region = defaultRegion
			log.Infof("NewDbConnexion Log: The AWS region for this environment has been set to: %s", cfg.Region)

	// We get the value for the API_ACCESS_TOKEN
		apiAccessToken, ok := os.LookupEnv("API_ACCESS_TOKEN")
		if ok {
			log.Infof("NewDbConnexion Log: API_ACCESS_TOKEN was overridden by local env: **hidden secret**")
		} else {
			log.Fatal("NewDbConnexion Fatal: API_ACCESS_TOKEN is unset as an environment variable, this is a fatal problem")
		}

	e, err := NewConfig(cfg)
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

	if err := http.ListenAndServe(addr, Protect(app, h.APIAccessToken)); err != nil {
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
