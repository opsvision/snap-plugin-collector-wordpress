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
package wordpress

import (
	"fmt"
	"sync"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

const (
	testing       = "http://opsvision.com" // test site
	pluginVendor  = "opsvision"            // vendor
	pluginName    = "wordpress"            // name
	pluginVersion = 1                      // version
)

type WordPress struct {
	Initialized  bool
	Connected    bool
	Availability uint64  // site availability 0 or 100
	ResolveTime  float64 // time to resolve site name in seconds
	ConnectTime  float64 // time to connect to site in seconds
	PageLoadTime float64 // time to download content in seconds
	TotalTime    float64 // total time to complete
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

func (w *WordPress) CollectMetrics(mts []plugin.Metric) ([]plugin.Metric, error) {

	/** Testing **/

	var wg sync.WaitGroup
	metrics := new(Metrics)

	// Get all the pages via REST API
	pages, err := GetPages(testing)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return nil, nil
	}

	// Set the WaitGroup size
	wg.Add(len(pages))

	// Iterate over the pages to get metrics
	for _, page := range pages {
		go func(page Page, metrics *Metrics) {
			defer wg.Done()
			var mutex sync.Mutex
			m := page.GetPageMetrics()
			mutex.Lock()
			*metrics = append(*metrics, m)
			mutex.Unlock()

		}(page, metrics)
	}

	wg.Wait()

	for _, m := range *metrics {
		fmt.Printf("%s | Page Load: %.6f | Resource Load: %.6f | Total Load: %.6f\n",
			m.Page, m.PageLoad, m.ResourceLoad, m.TotalLoad)
	}

	/** End Testing **/

	return nil, nil
}

func (w *WordPress) GetMetricTypes(cfg plugin.Config) ([]plugin.Metric, error) {
	return nil, nil
}

func (w *WordPress) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()

	return *policy, nil
}
