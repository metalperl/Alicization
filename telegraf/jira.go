package jira

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	jsonparser "github.com/influxdata/telegraf/plugins/parsers/json"
)

type Jira struct {
	Server          []string `toml:"server"`
	Project         string   `toml:"project"`
	Username        string   `toml:"username"`
	Password        string   `toml:"password"`
	GatherWeekly    bool     `toml:"gather_weekly"`
	GatherBiWeekly  bool     `toml:"gather_biweekly"`
	GatherMonthly   bool     `toml:"gather_monthly"`
	GatherQuarterly bool     `toml:"gather_quarterly"`
	GatherYearly    bool     `toml:"gather_yearly"`

	Timeout internal.Duration
	parser  parsers.Parser
}

var sampleConfig = `
  ## This plugin will query supplied project in jira
  ## Jira server to connect to
  ##  [protocol://[(hostname)]]
  ##  e.g.
  ##    https://jira.com
  ##
  ## if no servers are specified, local machine will be queried
  ##
  server = ["tcp(127.0.0.1::8080)/"]

  ## JIRA Project.
  project = ""

  ## HTTP Basic Authentication username and password.
  username = ""
  password = ""

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "json"

  ## gather metrics from issues within the last week for the jira server provided above
  gather_weekly    = false
  #
  ## gather metrics from issues within the last 2 weeks for the jira server provided above
  gather_biweekly  = true
  #
  ## gather metrics from issues within the month for the jira server provided above
  gather_monthly   = true
  #
  ## gather metrics from issues opened within the quarter for the jira server provided above
  gather_quarterly = false
  #
  ## gather metrics from issues opened within the year for the jira server provided above
  gather_yearly    = false
`

func (j *Jira) SampleConfig() string {
	return sampleConfig
}

func (j *Jira) Description() string {
	return "Read Jira given url and return project specified count for open and closed jira's"
}

func (j *Jira) SetParser(p parsers.Parser) {
	j.parser = p
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

var localhost = ""

func (j *Jira) Gather(acc telegraf.Accumulator) error {
	if len(j.Servers) == 0 {
		// default to localhost if nothing specified.
		return j.gatherMetrics(localhost, acc)
	}

	var wg sync.WaitGroup

	// Loop through each server and collect metrics
	for _, server := range j.Servers {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			acc.AddError(j.gatherMetrics(s, acc))
		}(server)
	}

	wg.Wait()
	return nil
}

func (j *Jira) gatherMetrics(serv string, acc telegraf.Accumulator) error {
	if j.GatherWeekly {
		rptType := "W"
		weekstart, weekend := weekly()
		rptStart := weekstart.Format("2006-01-02")
		rptEnd := weekend.Format("2006-01-02")
		err = j.buildJqlQuery(rptType, rptStart, rptEnd, serv, acc)
		if err != nil {
			return err
		}
	}

	if j.GatherBiWeekly {
		rptType := "B"
		biweekstart, biweekend := biweekly()
		rptStart := biweekstart.Format("2006-01-02")
		rptEnd := biweekend.Format("2006-01-02")
		err = j.buildJqlQuery(rptType, rptStart, rptEnd, serv, acc)
		if err != nil {
			return err
		}
	}

	if j.GatherMonthly {
		rptType := "M"
		monthstart, monthend := monthly()
		rptStart := monthstart.Format("2006-01-02")
		rptEnd := monthend.Format("2006-01-02")
		err = j.buildJqlQuery(rptType, rptStart, rptEnd, serv, acc)
		if err != nil {
			return err
		}
	}

	if j.GatherQuarterly {
		rptType := "Q"
		quarterstart, quarterend := quarterly()
		rptStart := quarterstart.Format("2006-01-02")
		rptEnd := quarterend.Format("2006-01-02")
		err = j.buildJqlQuery(rptType, rptStart, rptEnd, serv, acc)
		if err != nil {
			return err
		}
	}

	if j.GatherYearly {
		rptType := "Y"
		yearstart, yearend := yearly()
		rptStart := yearstart.Format("2006-01-02")
		rptEnd := yearend.Format("2006-01-02")
		err = j.buildJqlQuery(rptType, rptStart, rptEnd, serv, acc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (j *Jira) buildJqlQuery(rptType string, rptStart string, rptEnd string, serv string, acc telegraf.Accumulator) error {
	opnd := 0
	clsd := 0
	for i := 0; i < 2; i++ {
		if i == 0 {
			jql := "project =" + Project + " AND createdDate >=" + rptStart + " AND createdDate <=" + rptEnd + " AND status IN('open', 'in progress', 'reopened', 'waiting for customer', 'waiting for assignment', 'pending vendor')"
			opnd = runJqlQuery(jql, rptStart, rptEnd, serv)
		} else {
			jql := "project =" + Project + " AND createdDate >=" + rptStart + " AND createdDate <=" + rptEnd + " AND status IN('resolved', 'closed')"
			clsd = runJqlQuery(jql, rptStart, rptEnd, serv)
		}
	}
	j.reportOut(opnd, clsd, Project, rptType, rptStart, rptEnd)
	return nil
}

func (j *Jira) reportOut(opnd int, clsd int, Project string, rptType string, rptStart string, rptEnd string) error {
	// send the datams to acc
	epoch := rptStart + "-" + rptEnd
	tags := map[string]string{"project": Project, "epoch": epoch}
	fields := map[string]interface{}{
		"opened_jiras": opnd,
		"closed_jiras": clsd,
	}

	if rptType == "W" {
		acc.AddFields("jira_weekly", fields, tags)
	} else if rptType == "B" {
		acc.AddFields("jira_biweekly", fields, tags)
	} else if rptType == "M" {
		acc.AddFields("jira_monthly", fields, tags)
	} else if rptType == "Q" {
		acc.AddFields("jira_quarterly", fields, tags)
	} else if rptType == "Y" {
		acc.AddFields("jira_yearly", fields, tags)
	}
	return nil
}

func weekly() (time.Time, time.Time) {
	mydate := time.Now()
	day := mydate.Weekday()
	daynum := int(day)
	if daynum == 0 {
		lastSun := mydate.AddDate(0, 0, -7)
		lastSat := mydate.AddDate(0, 0, -1)
		return lastSun, lastSat
	} else if daynum == 1 {
		lastSun := mydate.AddDate(0, 0, -8)
		lastSat := mydate.AddDate(0, 0, -2)
		return lastSun, lastSat
	} else if daynum == 2 {
		lastSun := mydate.AddDate(0, 0, -9)
		lastSat := mydate.AddDate(0, 0, -3)
		return lastSun, lastSat
	} else if daynum == 3 {
		lastSun := mydate.AddDate(0, 0, -10)
		lastSat := mydate.AddDate(0, 0, -4)
		return lastSun, lastSat
	} else if daynum == 4 {
		lastSun := mydate.AddDate(0, 0, -11)
		lastSat := mydate.AddDate(0, 0, -5)
		return lastSun, lastSat
	} else if daynum == 5 {
		lastSun := mydate.AddDate(0, 0, -12)
		lastSat := mydate.AddDate(0, 0, -6)
		return lastSun, lastSat
	} else {
		lastSun := mydate.AddDate(0, 0, -13)
		lastSat := mydate.AddDate(0, 0, -7)
		return lastSun, lastSat
	}
}

func biweekly() (time.Time, time.Time) {
	mydate := time.Now()
	day := mydate.Weekday()
	daynum := int(day)
	if daynum == 0 {
		biweekSun := mydate.AddDate(0, 0, -14)
		lastSat := mydate.AddDate(0, 0, -1)
		return biweekSun, lastSat
	} else if daynum == 1 {
		biweekSun := mydate.AddDate(0, 0, -15)
		lastSat := mydate.AddDate(0, 0, -2)
		return biweekSun, lastSat
	} else if daynum == 2 {
		biweekSun := mydate.AddDate(0, 0, -16)
		lastSat := mydate.AddDate(0, 0, -3)
		return biweekSun, lastSat
	} else if daynum == 3 {
		biweekSun := mydate.AddDate(0, 0, -17)
		lastSat := mydate.AddDate(0, 0, -4)
		return biweekSun, lastSat
	} else if daynum == 4 {
		biweekSun := mydate.AddDate(0, 0, -18)
		lastSat := mydate.AddDate(0, 0, -5)
		return biweekSun, lastSat
	} else if daynum == 5 {
		biweekSun := mydate.AddDate(0, 0, -19)
		lastSat := mydate.AddDate(0, 0, -6)
		return biweekSun, lastSat
	} else {
		biweekSun := mydate.AddDate(0, 0, -20)
		lastSat := mydate.AddDate(0, 0, -7)
		return biweekSun, lastSat
	}
}

func monthly() (time.Time, time.Time) {
	mydate := time.Now()
	currentYear, currentMonth, _ := mydate.Date()
	currentLocation := mydate.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
	return firstOfMonth, lastOfMonth
}

func quarterly() (time.Time, time.Time) {
	mydate := time.Now()
	currentYear, currentMonth, _ := mydate.Date()
	currentLocation := mydate.Location()
	if currentMonth <= 3 {
		quarterStart := time.Date(currentYear, 1, 1, 0, 0, 0, 0, currentLocation)
		quarterEnd := quarterStart.AddDate(0, 3, -1)
		return quarterStart, quarterEnd
	} else if (currentMonth > 3) && (currentMonth <= 6) {
		quarterStart := time.Date(currentYear, 4, 1, 0, 0, 0, 0, currentLocation)
		quarterEnd := quarterStart.AddDate(0, 3, -1)
		return quarterStart, quarterEnd
	} else if (currentMonth > 6) && (currentMonth <= 9) {
		quarterStart := time.Date(currentYear, 7, 1, 0, 0, 0, 0, currentLocation)
		quarterEnd := quarterStart.AddDate(0, 3, -1)
		return quarterStart, quarterEnd
	} else {
		quarterStart := time.Date(currentYear, 10, 1, 0, 0, 0, 0, currentLocation)
		quarterEnd := quarterStart.AddDate(0, 3, -1)
		return quarterStart, quarterEnd
	}
}

func yearly() (time.Time, time.Time) {
	mydate := time.Now()
	currentYear, _, _ := mydate.Date()
	currentLocation := mydate.Location()
	firstOfYear := time.Date(currentYear, 1, 1, 0, 0, 0, 0, currentLocation)
	lastOfYear := firstOfYear.AddDate(0, 12, -1)
	return firstOfYear, lastOfYear
}

func runJqlQuery(jql string, rptStart string, rptEnd string, serv string) int {
	// Create the authenticated HTTP request
	client := &http.Client{Timeout: time.Second * 10}
	params := url.Values{}
	params.Add("jql", jql)
	req, err := http.NewRequest("GET", serv+"/rest/api/2/search?"+params.Encode(), nil)
	req.SetBasicAuth(Username, Password)
	resp, err := client.Do(req)
	checkError(err)

	// Read and parse JSON body
	defer resp.Body.Close()
	rawBody, err := ioutil.ReadAll(resp.Body)
	checkError(err)
	var jsonResult interface{}
	err = json.Unmarshal(rawBody, &jsonResult)
	checkError(err)
	m := jsonResult.(map[string]interface{})

	// extract the interesting data
	count := int(m["total"].(float64))

	// retun the datams
	return count
}

func init() {
	inputs.Add("jira", func() telegraf.Input {
		return &Jira{}
	})
}
