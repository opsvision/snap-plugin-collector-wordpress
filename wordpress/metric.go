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

// Metric contains load times in milliseconds
type Metric struct {
	Page         string  `json:"page"`
	PageLoad     float64 `json:"page_load"`
	ResourceLoad float64 `json:"resource_load"`
	TotalLoad    float64 `json:"total_load"`
}

// Metrics is a collection of Metric objects
type Metrics []Metric
