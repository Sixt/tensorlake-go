// Copyright 2025 SIXT SE
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tensorlake

import (
	"slices"
	"testing"
)

func TestDataset(t *testing.T) {
	c := initializeTestClient(t)

	// Create dataset.
	ds, err := c.CreateDataset(t.Context(), &CreateDatasetRequest{
		Name: "test_dataset",
		ParsingOptions: &ParsingOptions{
			ChunkingStrategy: ChunkingStrategyNone,
		},
		EnrichmentOptions: &EnrichmentOptions{
			TableSummarization: true,
		},
	})
	if err != nil {
		t.Fatalf("failed to create dataset: %v", err)
	}
	t.Logf("dataset created: %+v", ds)

	// List datasets.
	datasets := fetchAllDatasets(t, c)
	t.Logf("listed %d datasets: %v", len(datasets), datasets)

	// Check if the created dataset is in the list.
	if !slices.Contains(datasets, ds.DatasetId) {
		t.Fatalf("dataset %s not found in list", ds.DatasetId)
	}

	// Get dataset.
	dd, err := c.GetDataset(t.Context(), ds.DatasetId)
	if err != nil {
		t.Fatalf("failed to get dataset: %v", err)
	}
	t.Logf("dataset: %+v", dd)

	// Check if the dataset is the same as the created dataset.
	if dd.Name != ds.Name {
		t.Fatalf("dataset name mismatch: %s != %s", dd.Name, ds.Name)
	}

	// Update dataset: use page chunking strategy.
	dd, err = c.UpdateDataset(t.Context(), &UpdateDatasetRequest{
		DatasetId: ds.DatasetId,
		ParsingOptions: &ParsingOptions{
			ChunkingStrategy: ChunkingStrategyPage,
		},
	})
	if err != nil {
		t.Fatalf("failed to update dataset: %v", err)
	}
	t.Logf("dataset updated: %+v", dd)

	// Parse dataset.
	p, err := c.ParseDataset(t.Context(), &ParseDatasetRequest{
		DatasetId: ds.DatasetId,
		FileSource: FileSource{
			FileURL: "https://www.sixt.de/shared/t-c/sixt_DE_de.pdf",
		},
	})
	if err != nil {
		t.Fatalf("failed to parse dataset: %v", err)
	}
	t.Logf("dataset parse job: %+v", p)

	// Get parse job results.
	r, err := c.GetParseResult(t.Context(), p.ParseId, WithSSE(true), WithOnUpdate(func(eventName string, _ *ParseResult) {
		t.Logf("parse status: %s", eventName)
	}))
	if err != nil {
		t.Fatalf("failed to get parse job result: %v", err)
	}
	t.Logf("parse job result: %+v", r)

	// Delete parse job.
	err = c.DeleteParseJob(t.Context(), p.ParseId)
	if err != nil {
		t.Fatalf("failed to delete parse job: %v", err)
	}
	t.Logf("parse job deleted")

	// Delete dataset.
	err = c.DeleteDataset(t.Context(), ds.DatasetId)
	if err != nil {
		t.Fatalf("failed to delete dataset: %v", err)
	}
	t.Logf("dataset deleted")

	// Check if the dataset is deleted.
	datasets = fetchAllDatasets(t, c)
	t.Logf("listed %d datasets: %v", len(datasets), datasets)
	if slices.Contains(datasets, ds.DatasetId) {
		t.Fatalf("dataset %s is not deleted and found in list: %v", ds.DatasetId, datasets)
	}
}
