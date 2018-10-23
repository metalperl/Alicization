package cron_kpis

import (
	"bufio"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

)

type Cronkpis struct {
	Location string   `toml:"location"`
	Cronjob  string   `toml:"cron_job"`
	Host     string   `toml:"host"`
	CronCnt  []string `toml:"cron_count"`
}

var sampleConfig = `
  ## Location of the cron file log
  ##   For example:
  ##    "/home/naveen/go/src/filehandling/test.txt"
  location   = "/var/log/syslog"

  ## The name of the cron job to target
  ##   For example:
  ##     "WatchDogTimer.check"
  cron_job   = "WatchDogTimer.check"

  ## Host where the Cron jobs are running
  host       = "droozy-den-1p"

  ## Cron Count, the count of unique jobs that run for each day
  ##    jobs that run multiple times in a day are still only counted once
  ##    counts are listed starting with Sunday and need to be input for each day
  cron_count = ["3","4","5","6","7","0","0"]
`

func (c *Cronkpis) SampleConfig() string {
	return sampleConfig
}

func (c *Cronkpis) Description() string {
	return "Create kpi's from Cron Job statuses"
}

func (c *Cronkpis) Gather(acc telegraf.Accumulator) error {

	var wg sync.WaitGroup

	if len(c.Location) == 0 {
		return c.gatherStatuses(localdir, acc)
	} else {
		return c.gatherStatuses(c.Location, acc)
	}

	wg.Wait()
	return nil
}

func (c *Cronkpis) gatherStatuses(location string, acc telegraf.Accumulator) error {
	var (
		err     error
	)

	mydate := time.Now()
	day := mydate.Weekday()
	daynum := int(day)
	dater := mydate.Format("Jan 2")
	//t := time.Now()
	//p := t.Format(time.Kitchen)

	cronjob_exp := string(c.Cronjob)

	Go_NoGo := "0"

	//if ((p == "1:30AM") || (p == "1:31AM") || (p == "1:32AM") || (p == "1:33AM")) {

		datam, err := os.Open(location)
		if err != nil {
			return err
		}
		defer datam.Close()

		// Splits on newlines by default.
		scanner := bufio.NewScanner(datam)

		for scanner.Scan() {
			if strings.Contains(scanner.Text(), cronjob_exp) {
				if strings.Contains(scanner.Text(), dater) {
					if daynum == 0 {
						Go_NoGo = (string(c.CronCnt[0]))
						Go_NoGo = strings.TrimSpace(Go_NoGo)
					} else if daynum == 5 {
						Go_NoGo = (string(c.CronCnt[5]))
						Go_NoGo = strings.TrimSpace(Go_NoGo)
					} else if daynum == 6 {
						Go_NoGo = (string(c.CronCnt[6]))
						Go_NoGo = strings.TrimSpace(Go_NoGo)
					} else {
						Go_NoGo = (string(c.CronCnt[1]))
						Go_NoGo = strings.TrimSpace(Go_NoGo)
					}
				}
			}
		}

		// cron_kpis map[cron_on_off:1] map[host:itrccog-wc-1p]
		tags := map[string]string{"server": c.Host}
		//fields := make(map[string]interface{})

		//fields["cron_count"] = Go_NoGo
		fields := map[string]interface{}{"cron_count": Go_NoGo,}
		acc.AddFields("cron_kpis", fields, tags)

	//}

	return nil
}

func (c *Cronkpis) accumulateCron(Go_NoGo string, Host string, acc telegraf.Accumulator) error {
	cron_cnt  := Go_NoGo
	host_name := Host
	// cron_kpis map[cron_count:1] map[hostname:itrccog-wc-1p]
	tags := map[string]string{"hostname": host_name}
	fields := make(map[string]interface{})

	fields["cron_count"] = cron_cnt
	acc.AddFields("cron_kpis", fields, tags)

	return nil
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

var localdir = ""

func init() {
	inputs.Add("cron_kpis", func() telegraf.Input {
		return &Cronkpis{
		}
	})
}