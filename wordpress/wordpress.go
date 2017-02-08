package wordpress

/*
 * http://www.apache.org/licenses/LICENSE-2.0.txt
 *
 * Copyright 2017 OpsVision Solutions
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

const (
	testing       = "http://opsvision.com" // test site
	pluginVendor  = "opsvision"            // vendor
	pluginName    = "wordpress"            // name
	pluginVersion = 1                      // version
)

// WordPress is our client object
type WordPress struct {
	Initialized bool
}

// New instantiates our client
func New() *WordPress {
	return new(WordPress)
}

// init is used to initialize our client
func (w *WordPress) init(cfg plugin.Config) {
	if w.Initialized {
		return
	}

	w.Initialized = true
}

// CollectMetrics is called by Snap-Telemetry to gather metrics
func (w *WordPress) CollectMetrics(mts []plugin.Metric) ([]plugin.Metric, error) {

	/** Testing **/

	var wg sync.WaitGroup
	var mutex sync.Mutex

	re := regexp.MustCompile("^[HTTPhttp]+[Ss]?://")
	site := re.ReplaceAllString(testing, "")
	metrics := new(Metrics)
	avail := 100

	// Get all the pages via REST API
	pages, err := GetPages(testing)
	if err != nil || len(pages) == 0 {
		//fmt.Printf("Error: %s\n", err.Error())
		avail = 0

	}

	// Set the WaitGroup size
	wg.Add(len(pages))

	// Iterate over the pages to get metrics
	for _, page := range pages {
		go func(page Page, metrics *Metrics) {
			defer wg.Done()

			// Get a mutex lock
			mutex.Lock()
			defer mutex.Unlock()

			// Store the metrics
			*metrics = append(*metrics, page.GetPageMetrics())
		}(page, metrics)
	}

	wg.Wait()

	// Site Availability
	fmt.Printf("/%s/%s/availability: %d\n", pluginVendor, site, avail)

	for _, m := range *metrics {
		fmt.Printf("/%s/%s/%s/page_load: %.6f\n",
			pluginVendor, site, m.Page, m.PageLoad)

		fmt.Printf("/%s/%s/%s/resource_load: %.6f\n",
			pluginVendor, site, m.Page, m.ResourceLoad)

		fmt.Printf("/%s/%s/%s/total_load: %.6f\n",
			pluginVendor, site, m.Page, m.TotalLoad)
	}

	/** End Testing **/

	return nil, nil
}

// GetMetricTypes returns metric types for testing
func (w *WordPress) GetMetricTypes(cfg plugin.Config) ([]plugin.Metric, error) {
	return nil, nil
}

// GetConfigPolicy returns the configPolicy for your plugin
func (w *WordPress) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()

	return *policy, nil
}
