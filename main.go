package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	jsonhandler "github.com/apex/log/handlers/json"
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/gorilla/pat"
	"github.com/tj/go/http/response"
	"github.com/unee-t/env"

	"github.com/apex/log"
	"github.com/apex/log/handlers/text"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type handler struct {
	DSN            string // e.g. "bugzilla:secret@tcp(auroradb.dev.unee-t.com:3306)/bugzilla?multiStatements=true&sql_mode=TRADITIONAL"
	Domain         string // e.g. https://dev.case.unee-t.com
	APIAccessToken string // e.g. O8I9svDTizOfLfdVA5ri
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

// New setups the configuration assuming various parameters have been setup in the AWS account
func New() (h handler, err error) {

	cfg, err := external.LoadDefaultAWSConfig(external.WithSharedConfigProfile("uneet-dev"))
	if err != nil {
		log.WithError(err).Fatal("setting up credentials")
		return
	}
	cfg.Region = endpoints.ApSoutheast1RegionID
	e, err := env.New(cfg)
	if err != nil {
		log.WithError(err).Fatal("error getting unee-t env")
	}

	h = handler{
		DSN: fmt.Sprintf("bugzilla:%s@tcp(%s:3306)/bugzilla?multiStatements=true&sql_mode=TRADITIONAL",
			e.GetSecret("MYSQL_PASSWORD"),
			e.Udomain("auroradb")),
		Domain:         fmt.Sprintf("https://%s", e.Udomain("case")),
		APIAccessToken: e.GetSecret("API_ACCESS_TOKEN"),
		Code:           e.Code,
	}

	if h.Code == 0 {
		err = fmt.Errorf("Error code is unknown/unset")
		return
	}

	log.Infof("Frontend URL: %v", h.Domain)

	h.db, err = sql.Open("mysql", h.DSN)
	if err != nil {
		log.WithError(err).Fatal("error opening database")
		return
	}

	return

}

func main() {

	h, err := New()
	if err != nil {
		log.WithError(err).Fatal("error setting configuration")
		return
	}

	defer h.db.Close()

	addr := ":" + os.Getenv("PORT")
	app := pat.New()
	app.Post("/", http.HandlerFunc(h.enroll))
	if err := http.ListenAndServe(addr, app); err != nil {
		log.WithError(err).Fatal("error listening")
	}

}

func (h handler) enroll(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	var k APIkey
	err := decoder.Decode(&k)

	if err != nil {
		log.WithError(err).Errorf("Input error")
		response.BadRequest(w, "Invalid JSON")
		return
	}
	defer r.Body.Close()

	ctx := log.WithFields(log.Fields{
		"APIkey": k,
	})

	ctx.Info("Decoded")

	if k.UserAPIkey == "" {
		response.BadRequest(w, "Missing UserAPIkey")
		return
	}

	if k.UserID == "" {
		response.BadRequest(w, "Missing UserID")
		return
	}

	response.OK(w)

}
