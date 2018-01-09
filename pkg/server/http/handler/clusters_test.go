package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/apprenda/kismatic/pkg/store"
	"github.com/julienschmidt/httprouter"
)

type mockClustersStore struct {
	store map[string]store.Cluster
}

func (cs mockClustersStore) Get(key string) (*store.Cluster, error) {
	c, ok := cs.store[key]
	if !ok {
		return nil, nil
	}
	return &c, nil
}
func (cs *mockClustersStore) Put(key string, cluster store.Cluster) error {
	if cs.store == nil {
		cs.store = make(map[string]store.Cluster)
	}
	cs.store[key] = cluster
	return nil
}

func (cs mockClustersStore) GetAll() (map[string]store.Cluster, error) {
	return cs.store, nil
}

func (cs mockClustersStore) Delete(key string) error {
	delete(cs.store, key)
	return nil
}

func (cs mockClustersStore) Watch(ctx context.Context, buffer uint) <-chan store.WatchResponse {
	return nil
}

func TestValidationShouldError(t *testing.T) {
	if testing.Short() {
		return
	}
	tests := []*ClusterRequest{
		&ClusterRequest{
			Name:         "",
			DesiredState: "installed",
			Provisioner: Provisioner{
				Provider: "aws",
				Options: map[string]string{
					awsOptionAccessKeyID:     "ACCESS_ID",
					awsOptionSecretAccessKey: "SECRET",
				},
			},
			EtcdCount:    3,
			MasterCount:  2,
			WorkerCount:  5,
			IngressCount: 2,
		},
		&ClusterRequest{
			Name:         "foo",
			DesiredState: "bar",
			Provisioner: Provisioner{
				Provider: "aws",
				Options: map[string]string{
					awsOptionAccessKeyID:     "ACCESS_ID",
					awsOptionSecretAccessKey: "SECRET",
				},
			},
			EtcdCount:    3,
			MasterCount:  2,
			WorkerCount:  5,
			IngressCount: 2,
		},
		&ClusterRequest{
			Name:         "foo",
			DesiredState: "installed",
			Provisioner: Provisioner{
				Provider: "foobar",
				Options: map[string]string{
					awsOptionAccessKeyID:     "ACCESS_ID",
					awsOptionSecretAccessKey: "SECRET",
				},
			},
			EtcdCount:    3,
			MasterCount:  2,
			WorkerCount:  5,
			IngressCount: 2,
		},
		&ClusterRequest{
			Name:         "foo",
			DesiredState: "installed",
			Provisioner: Provisioner{
				Provider: "aws",
				Options: map[string]string{
					awsOptionAccessKeyID:     "",
					awsOptionSecretAccessKey: "SECRET",
				},
			},
			EtcdCount:    3,
			MasterCount:  2,
			WorkerCount:  5,
			IngressCount: 2,
		},
		&ClusterRequest{
			Name:         "foo",
			DesiredState: "installed",
			Provisioner: Provisioner{
				Provider: "aws",
				Options: map[string]string{
					awsOptionAccessKeyID:     "ACCESS_ID",
					awsOptionSecretAccessKey: "",
				},
			},
			EtcdCount:    3,
			MasterCount:  2,
			WorkerCount:  5,
			IngressCount: 2,
		},
		&ClusterRequest{
			Name:         "foo",
			DesiredState: "installed",
			Provisioner: Provisioner{
				Provider: "aws",
				Options: map[string]string{
					awsOptionAccessKeyID:     "ACCESS_ID",
					awsOptionSecretAccessKey: "SECRET",
				},
			},
			EtcdCount:    0,
			MasterCount:  2,
			WorkerCount:  5,
			IngressCount: 2,
		},
		&ClusterRequest{
			Name:         "foo",
			DesiredState: "installed",
			Provisioner: Provisioner{
				Provider: "aws",
				Options: map[string]string{
					awsOptionAccessKeyID:     "ACCESS_ID",
					awsOptionSecretAccessKey: "SECRET",
				},
			},
			EtcdCount:    3,
			MasterCount:  0,
			WorkerCount:  5,
			IngressCount: 2,
		},
		&ClusterRequest{
			Name:         "foo",
			DesiredState: "installed",
			Provisioner: Provisioner{
				Provider: "aws",
				Options: map[string]string{
					awsOptionAccessKeyID:     "ACCESS_ID",
					awsOptionSecretAccessKey: "SECRET",
				},
			},
			EtcdCount:    3,
			MasterCount:  2,
			WorkerCount:  0,
			IngressCount: 2,
		},
		&ClusterRequest{
			Name:         "foo",
			DesiredState: "installed",
			Provisioner: Provisioner{
				Provider: "aws",
				Options: map[string]string{
					awsOptionAccessKeyID:     "ACCESS_ID",
					awsOptionSecretAccessKey: "SECRET",
				},
			},
			EtcdCount:    3,
			MasterCount:  2,
			WorkerCount:  5,
			IngressCount: -1,
		},
	}
	for i, c := range tests {
		encoded, err := json.Marshal(c)
		if err != nil {
			t.Fatalf("could not encode body to json %v", err)
		}
		// Create a request to pass to our handler
		req, err := http.NewRequest("POST", "/clusters", bytes.NewBuffer(encoded))
		if err != nil {
			t.Fatal(err)
		}
		// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response
		rr := httptest.NewRecorder()

		// Call their ServeHTTP method directly and pass in our Request and ResponseRecorder
		r := httprouter.New()

		cs := &mockClustersStore{}
		clustersAPI := Clusters{Store: cs, Logger: log.New(os.Stdout, "test", 0)}
		r.POST("/clusters", clustersAPI.Create)
		r.ServeHTTP(rr, req)

		// Check the status code is as expected
		if status := rr.Code; status != http.StatusBadRequest {
			t.Logf("running test: %d", i)
			t.Errorf("handler returned wrong status code: got %v want %v: %s",
				status, http.StatusBadRequest, rr.Body.String())
		}
	}
}

func TestUpdateValidationShouldError(t *testing.T) {
	tests := []struct {
		name  string
		cu    clusterUpdate
		valid bool
	}{
		{
			name: "updating with no changes is valid",
			cu: clusterUpdate{
				id: "foo",
				request: ClusterRequest{
					Name:         "foo",
					DesiredState: "installed",
					Provisioner: Provisioner{
						Provider: "aws",
						Options: map[string]string{
							awsOptionRegion:          "us-east-1",
							awsOptionAccessKeyID:     "ACCESS_ID",
							awsOptionSecretAccessKey: "SECRET",
						},
					},
					EtcdCount:    3,
					MasterCount:  2,
					WorkerCount:  5,
					IngressCount: 2,
				},
				inStore: store.Cluster{
					Spec: store.ClusterSpec{
						EtcdCount:    3,
						MasterCount:  2,
						WorkerCount:  5,
						IngressCount: 2,
					},
				},
			},
			valid: true,
		},
		{
			name: "updating the number of ingress nodes is valid",
			cu: clusterUpdate{
				id: "foo",
				request: ClusterRequest{
					Name:         "foo",
					DesiredState: "installed",
					Provisioner: Provisioner{
						Provider: "aws",
						Options: map[string]string{
							awsOptionRegion:          "us-east-1",
							awsOptionAccessKeyID:     "ACCESS_ID",
							awsOptionSecretAccessKey: "SECRET",
						},
					},
					EtcdCount:    3,
					MasterCount:  2,
					WorkerCount:  5,
					IngressCount: 0,
				},
				inStore: store.Cluster{
					Spec: store.ClusterSpec{
						EtcdCount:    3,
						MasterCount:  2,
						WorkerCount:  5,
						IngressCount: 2,
					},
				},
			},
			valid: true,
		},
		{
			name: "updating the number of worker nodes is valid",
			cu: clusterUpdate{
				id: "foo",
				request: ClusterRequest{
					Name:         "foo",
					DesiredState: "installed",
					Provisioner: Provisioner{
						Provider: "aws",
						Options: map[string]string{
							awsOptionRegion:          "us-east-1",
							awsOptionAccessKeyID:     "ACCESS_ID_NEW",
							awsOptionSecretAccessKey: "SECRET_NEW",
						},
					},
					EtcdCount:    3,
					MasterCount:  3,
					WorkerCount:  6,
					IngressCount: 3,
				},
				inStore: store.Cluster{
					Spec: store.ClusterSpec{
						EtcdCount:    3,
						MasterCount:  2,
						WorkerCount:  5,
						IngressCount: 2,
					},
				},
			},
			valid: true,
		},
		{
			name: "updating the name of the cluster is not valid",
			cu: clusterUpdate{
				id: "bar",
				request: ClusterRequest{
					Name:         "foo",
					DesiredState: "installed",
					Provisioner: Provisioner{
						Provider: "aws",
						Options: map[string]string{
							awsOptionRegion:          "us-east-1",
							awsOptionAccessKeyID:     "ACCESS_ID",
							awsOptionSecretAccessKey: "SECRET",
						},
					},
					EtcdCount:    3,
					MasterCount:  2,
					WorkerCount:  5,
					IngressCount: 2,
				},
				inStore: store.Cluster{
					Spec: store.ClusterSpec{
						EtcdCount:    3,
						MasterCount:  2,
						WorkerCount:  5,
						IngressCount: 2,
					},
				},
			},
			valid: false,
		},
		{
			name: "updating the name of the cluster is not valid",
			cu: clusterUpdate{
				id: "foo",
				request: ClusterRequest{
					Name:         "bar",
					DesiredState: "installed",
					Provisioner: Provisioner{
						Provider: "aws",
						Options: map[string]string{
							awsOptionRegion:          "us-east-1",
							awsOptionAccessKeyID:     "ACCESS_ID",
							awsOptionSecretAccessKey: "SECRET",
						},
					},
					EtcdCount:    3,
					MasterCount:  2,
					WorkerCount:  5,
					IngressCount: 2,
				},
				inStore: store.Cluster{
					Spec: store.ClusterSpec{
						EtcdCount:    3,
						MasterCount:  2,
						WorkerCount:  5,
						IngressCount: 2,
					},
				},
			},
			valid: false,
		},
		{
			name: "updating the number of etcd nodes is not valid",
			cu: clusterUpdate{
				id: "foo",
				request: ClusterRequest{
					Name:         "foo",
					DesiredState: "installed",
					Provisioner: Provisioner{
						Provider: "aws",
						Options: map[string]string{
							awsOptionRegion:          "us-east-1",
							awsOptionAccessKeyID:     "ACCESS_ID",
							awsOptionSecretAccessKey: "SECRET",
						},
					},
					EtcdCount:    5,
					MasterCount:  2,
					WorkerCount:  5,
					IngressCount: 2,
				},
				inStore: store.Cluster{
					Spec: store.ClusterSpec{
						EtcdCount:    3,
						MasterCount:  2,
						WorkerCount:  5,
						IngressCount: 2,
					},
				},
			},
			valid: false,
		},
		{
			name: "updating master nodes to zero is not valid",
			cu: clusterUpdate{
				id: "foo",
				request: ClusterRequest{
					Name:         "foo",
					DesiredState: "installed",
					Provisioner: Provisioner{
						Provider: "aws",
						Options: map[string]string{
							awsOptionRegion:          "us-east-1",
							awsOptionAccessKeyID:     "ACCESS_ID",
							awsOptionSecretAccessKey: "SECRET",
						},
					},
					EtcdCount:    3,
					MasterCount:  0,
					WorkerCount:  5,
					IngressCount: 2,
				},
				inStore: store.Cluster{
					Spec: store.ClusterSpec{
						EtcdCount:    3,
						MasterCount:  2,
						WorkerCount:  5,
						IngressCount: 2,
					},
				},
			},
			valid: false,
		},
		{
			name: "updating the worker nodes to zero is not valid",
			cu: clusterUpdate{
				id: "foo",
				request: ClusterRequest{
					Name:         "foo",
					DesiredState: "installed",
					Provisioner: Provisioner{
						Provider: "aws",
						Options: map[string]string{
							awsOptionRegion:          "us-east-1",
							awsOptionAccessKeyID:     "ACCESS_ID",
							awsOptionSecretAccessKey: "SECRET",
						},
					},
					EtcdCount:    3,
					MasterCount:  2,
					WorkerCount:  0,
					IngressCount: 2,
				},
				inStore: store.Cluster{
					Spec: store.ClusterSpec{
						EtcdCount:    3,
						MasterCount:  2,
						WorkerCount:  5,
						IngressCount: 2,
					},
				},
			},
			valid: false,
		},
		{
			name: "updating the number of ingress nodes to < 0 is not valid",
			cu: clusterUpdate{
				id: "foo",
				request: ClusterRequest{
					Name:         "foo",
					DesiredState: "installed",
					Provisioner: Provisioner{
						Provider: "aws",
						Options: map[string]string{
							awsOptionRegion:          "us-east-1",
							awsOptionAccessKeyID:     "ACCESS_ID",
							awsOptionSecretAccessKey: "SECRET",
						},
					},
					EtcdCount:    3,
					MasterCount:  2,
					WorkerCount:  5,
					IngressCount: -1,
				},
				inStore: store.Cluster{
					Spec: store.ClusterSpec{
						EtcdCount:    3,
						MasterCount:  2,
						WorkerCount:  5,
						IngressCount: 2,
					},
				},
			},
			valid: false,
		},
	}
	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ok, errs := test.cu.validate()
			if ok != test.valid {
				t.Errorf("test %d: expect %t, but got %t: %v", i, test.valid, ok, errs)
			}
		})
	}
}

func TestValidation(t *testing.T) {
	if testing.Short() {
		return
	}
	tests := []*ClusterRequest{
		&ClusterRequest{
			Name:         "foo",
			DesiredState: "installed",
			Provisioner: Provisioner{
				Provider: "aws",
				Options: map[string]string{
					awsOptionRegion:          "us-east-1",
					awsOptionAccessKeyID:     "ACCESS_ID",
					awsOptionSecretAccessKey: "SECRET",
				},
			},
			EtcdCount:    3,
			MasterCount:  2,
			WorkerCount:  5,
			IngressCount: 2,
		},
	}
	for _, c := range tests {
		encoded, err := json.Marshal(c)
		if err != nil {
			t.Fatalf("could not encode body to json %v", err)
		}
		// Create a request to pass to our handler
		req, err := http.NewRequest("POST", "/clusters", bytes.NewBuffer(encoded))
		if err != nil {
			t.Fatal(err)
		}
		// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response
		rr := httptest.NewRecorder()

		// Call their ServeHTTP method directly and pass in our Request and ResponseRecorder
		r := httprouter.New()

		cs := &mockClustersStore{}
		clustersAPI := Clusters{Store: cs, Logger: log.New(os.Stdout, "test", 0)}
		r.POST("/clusters", clustersAPI.Create)
		r.ServeHTTP(rr, req)

		// Check the status code is as expected
		if status := rr.Code; status != http.StatusAccepted {
			t.Errorf("handler returned wrong status code: got %v want %v: %s",
				status, http.StatusAccepted, rr.Body.String())
		}
	}
}

func TestCreateUpdateGetGetAllandDelete(t *testing.T) {
	if testing.Short() {
		return
	}
	c := &ClusterRequest{
		Name:         "foo",
		DesiredState: "installed",
		Provisioner: Provisioner{
			Provider: "aws",
			Options: map[string]string{
				awsOptionRegion:          "us-east-1",
				awsOptionAccessKeyID:     "ACCESS_ID",
				awsOptionSecretAccessKey: "SECRET",
			},
		},
		EtcdCount:    3,
		MasterCount:  2,
		WorkerCount:  5,
		IngressCount: 2,
	}
	encoded, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("could not encode body to json %v", err)
	}
	// Create a request to pass to our handler
	req, err := http.NewRequest("POST", "/clusters", bytes.NewBuffer(encoded))
	if err != nil {
		t.Fatal(err)
	}
	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response
	rr := httptest.NewRecorder()

	// Call their ServeHTTP method directly and pass in our Request and ResponseRecorder
	r := httprouter.New()

	cs := &mockClustersStore{}
	clustersAPI := Clusters{Store: cs, Logger: log.New(os.Stdout, "test", 0)}
	r.POST("/clusters", clustersAPI.Create)
	r.ServeHTTP(rr, req)

	// Check the status code is as expected
	if status := rr.Code; status != http.StatusAccepted {
		t.Errorf("handler returned wrong status code: got %v want %v: %s",
			status, http.StatusAccepted, rr.Body.String())
	}

	// Check the response body is as expected
	expected := "ok\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}

	// should update
	c = &ClusterRequest{
		Name:         "foo",
		DesiredState: "installed",
		Provisioner: Provisioner{
			Provider: "aws",
			Options: map[string]string{
				awsOptionRegion:          "us-east-1",
				awsOptionAccessKeyID:     "ACCESS_ID",
				awsOptionSecretAccessKey: "SECRET",
			},
		},
		EtcdCount:    3,
		MasterCount:  2,
		WorkerCount:  6,
		IngressCount: 2,
	}
	encoded, err = json.Marshal(c)
	if err != nil {
		t.Fatalf("could not encode body to json %v", err)
	}
	// Create a request to pass to our handler
	req, err = http.NewRequest("PUT", "/clusters/foo", bytes.NewBuffer(encoded))
	if err != nil {
		t.Error(err)
	}
	rr = httptest.NewRecorder()
	r = httprouter.New()
	r.PUT("/clusters/:name", clustersAPI.Update)
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusAccepted {
		t.Errorf("handler returned wrong status code: got %v want %v: %s",
			status, http.StatusAccepted, rr.Body.String())
	}
	resp := &ClusterResponse{}
	err = json.NewDecoder(rr.Body).Decode(resp)
	if err != nil {
		t.Error(err)
	}
	if resp.WorkerCount != 6 {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), resp)
	}

	// should get 404
	req, err = http.NewRequest("GET", "/clusters/bar", nil)
	if err != nil {
		t.Error(err)
	}
	rr = httptest.NewRecorder()
	r = httprouter.New()
	r.GET("/clusters/:name", clustersAPI.Get)
	r.ServeHTTP(rr, req)
	// Check the status code is 404
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v: %s",
			status, http.StatusNotFound, rr.Body.String())
	}

	// should get a response
	req, err = http.NewRequest("GET", "/clusters/foo", nil)
	if err != nil {
		t.Error(err)
	}
	rr = httptest.NewRecorder()
	r = httprouter.New()
	r.GET("/clusters/:name", clustersAPI.Get)
	r.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v: %s",
			status, http.StatusOK, rr.Body.String())
	}

	// should getAll
	req, err = http.NewRequest("GET", "/clusters", nil)
	if err != nil {
		t.Error(err)
	}
	rr = httptest.NewRecorder()
	r = httprouter.New()
	r.GET("/clusters", clustersAPI.GetAll)
	r.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v: %s",
			status, http.StatusOK, rr.Body.String())
	}
	respAll := make([]ClusterResponse, 0)
	err = json.NewDecoder(rr.Body).Decode(&respAll)
	if err != nil {
		t.Error(err)
	}
	if len(respAll) != 1 {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), resp)
	}

	// should delete
	req, err = http.NewRequest("DELETE", "/clusters/foo", nil)
	if err != nil {
		t.Error(err)
	}
	rr = httptest.NewRecorder()
	r = httprouter.New()
	r.DELETE("/clusters/:name", clustersAPI.Delete)
	r.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusAccepted {
		t.Errorf("handler returned wrong status code: got %d want %v: %d",
			status, http.StatusAccepted, rr.Code)
	}
	expected = "ok\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}

	// should getAll
	req, err = http.NewRequest("GET", "/clusters", nil)
	if err != nil {
		t.Error(err)
	}
	rr = httptest.NewRecorder()
	r = httprouter.New()
	r.GET("/clusters", clustersAPI.GetAll)
	r.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v: %s",
			status, http.StatusOK, rr.Body.String())
	}
	respAll = make([]ClusterResponse, 0)
	err = json.NewDecoder(rr.Body).Decode(&respAll)
	if err != nil {
		t.Error(err)
	}
	if len(respAll) != 1 && respAll[0].DesiredState != "destroyed" {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), respAll)
	}
}

func TestProviderOptions(t *testing.T) {
	cs := &mockClustersStore{}
	handler := Clusters{Store: cs, Logger: log.New(os.Stdout, "test", 0)}
	clusterRequest := &ClusterRequest{
		Name:         "foo",
		DesiredState: "installed",
		Provisioner: Provisioner{
			Provider: "aws",
			Options: map[string]string{
				awsOptionAccessKeyID:     "ACCESS_ID",
				awsOptionSecretAccessKey: "SECRET",
				awsOptionRegion:          "us-east-2",
			},
		},
		EtcdCount:    3,
		MasterCount:  2,
		WorkerCount:  5,
		IngressCount: 2,
	}
	encoded, err := json.Marshal(clusterRequest)
	if err != nil {
		t.Fatalf("could not encode body to json %v", err)
	}
	req, err := http.NewRequest("POST", "/clusters", bytes.NewBuffer(encoded))
	if err != nil {
		t.Fatal(err)
	}
	handler.Create(httptest.NewRecorder(), req, httprouter.Params{})
	cluster, err := cs.Get(clusterRequest.Name)
	if err != nil {
		t.Fatalf("cluster was not persisted in the store")
	}
	if cluster.Spec.Provisioner.Options.AWS.Region != clusterRequest.Provisioner.Options[awsOptionRegion] {
		t.Errorf("AWS region was not persisted. Expected %s, but got %s", clusterRequest.Provisioner.Options[awsOptionRegion], cluster.Spec.Provisioner.Options.AWS.Region)
	}

	recorder := httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/clusters/"+clusterRequest.Name, nil)
	if err != nil {
		t.Fatalf("error building request: %v", err)
	}
	handler.Get(recorder, req, httprouter.Params{httprouter.Param{Key: "name", Value: clusterRequest.Name}})
	clusterResponse := ClusterResponse{}
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected response from the server: %d", recorder.Code)
	}
	err = json.NewDecoder(recorder.Body).Decode(&clusterResponse)
	if err != nil {
		t.Fatalf("error decoding response from server: %v", err)
	}
	if clusterResponse.Provisioner.Options[awsOptionRegion] != clusterRequest.Provisioner.Options[awsOptionRegion] {
		t.Errorf("cluster response has region %q, but request had region %q", clusterResponse.Provisioner.Options[awsOptionRegion], clusterRequest.Provisioner.Options[awsOptionRegion])
	}
}

func TestGetKubeconfig(t *testing.T) {
	cs := &mockClustersStore{}
	cs.Put("foo", store.Cluster{})
	cs.Put("foobar", store.Cluster{})

	// Call their ServeHTTP method directly and pass in our Request and ResponseRecorder
	r := httprouter.New()

	assetsDir, err := mockAssetsDir()
	if err != nil {
		t.Fatal(err)
	}

	clustersAPI := Clusters{Store: cs, AssetsDir: assetsDir, Logger: log.New(os.Stdout, "test", 0)}
	r.GET("/clusters/:name/kubeconfig", clustersAPI.GetKubeconfig)

	// Create a request to pass to our handler
	req, err := http.NewRequest("GET", "/clusters/foo/kubeconfig", nil)
	if err != nil {
		t.Error(err)
	}
	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v: %s",
			status, http.StatusOK, rr.Body.String())
	}
	if strings.TrimSpace(rr.Body.String()) != "kubeconfig" {
		t.Errorf("response was not what was expecteded\ngot: %v\nexpected: %v", rr.Body.String(), "kubeconfig")
	}

	// Create a request to pass to our handler that should return a 404
	req, err = http.NewRequest("GET", "/clusters/bar/kubeconfig", nil)
	if err != nil {
		t.Error(err)
	}
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v: %s",
			status, http.StatusNotFound, rr.Body.String())
	}

	// Create a request to pass to our handler that should return a 500
	// Exists in store but not in the assets dir
	req, err = http.NewRequest("GET", "/clusters/foobar/kubeconfig", nil)
	if err != nil {
		t.Error(err)
	}
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v: %s",
			status, http.StatusInternalServerError, rr.Body.String())
	}
}

func TestGetLogs(t *testing.T) {
	cs := &mockClustersStore{}
	cs.Put("foo", store.Cluster{})
	cs.Put("foobar", store.Cluster{})

	// Call their ServeHTTP method directly and pass in our Request and ResponseRecorder
	r := httprouter.New()

	assetsDir, err := mockAssetsDir()
	if err != nil {
		t.Fatal(err)
	}

	clustersAPI := Clusters{Store: cs, AssetsDir: assetsDir, Logger: log.New(os.Stdout, "test", 0)}
	r.GET("/clusters/:name/logs", clustersAPI.GetLogs)

	// Create a request to pass to our handler
	req, err := http.NewRequest("GET", "/clusters/foo/logs", nil)
	if err != nil {
		t.Error(err)
	}
	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v: %s",
			status, http.StatusOK, rr.Body.String())
	}
	if strings.TrimSpace(rr.Body.String()) != "logs" {
		t.Errorf("response was not what was expecteded\ngot: %v\nexpected: %v", rr.Body.String(), "logs")
	}

	// Create a request to pass to our handler that should return a 404
	req, err = http.NewRequest("GET", "/clusters/bar/logs", nil)
	if err != nil {
		t.Error(err)
	}
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v: %s",
			status, http.StatusNotFound, rr.Body.String())
	}

	// Create a request to pass to our handler that should return a 500
	// Exists in store but not in the assets dir
	req, err = http.NewRequest("GET", "/clusters/foobar/logs", nil)
	if err != nil {
		t.Error(err)
	}
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v: %s",
			status, http.StatusInternalServerError, rr.Body.String())
	}
}

func TestGetAssets(t *testing.T) {
	cs := &mockClustersStore{}
	cs.Put("foo", store.Cluster{})
	cs.Put("foobar", store.Cluster{})

	// Call their ServeHTTP method directly and pass in our Request and ResponseRecorder
	r := httprouter.New()

	assetsDir, err := mockAssetsDir()
	if err != nil {
		t.Fatal(err)
	}

	clustersAPI := Clusters{Store: cs, AssetsDir: assetsDir, Logger: log.New(os.Stdout, "test", 0)}
	r.GET("/clusters/:name/assets", clustersAPI.GetKubeconfig)

	// Create a request to pass to our handler
	req, err := http.NewRequest("GET", "/clusters/foo/assets", nil)
	if err != nil {
		t.Error(err)
	}
	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v: %s",
			status, http.StatusOK, rr.Body.String())
	}

	// Create a request to pass to our handler that should return a 404
	req, err = http.NewRequest("GET", "/clusters/bar/assets", nil)
	if err != nil {
		t.Error(err)
	}
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v: %s",
			status, http.StatusNotFound, rr.Body.String())
	}

	// Create a request to pass to our handler that should return a 500
	// Exists in store but not in the assets dir
	req, err = http.NewRequest("GET", "/clusters/foobar/assets", nil)
	if err != nil {
		t.Error(err)
	}
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v: %s",
			status, http.StatusInternalServerError, rr.Body.String())
	}
}

func mockAssetsDir() (string, error) {
	assetsDir, err := ioutil.TempDir("/tmp", "ket-server-assets")
	if err != nil {
		return "", fmt.Errorf("error creating assets directory %q: %v", assetsDir, err)
	}

	generatedDir := path.Join(assetsDir, "foo", "assets")
	err = os.MkdirAll(generatedDir, 0770)
	if err != nil {
		return "", fmt.Errorf("Error creating generated directory %q: %v", generatedDir, err)
	}

	// write a fake kubeconfig file
	configd := []byte("kubeconfig")
	err = ioutil.WriteFile(path.Join(generatedDir, "kubeconfig"), configd, 0644)
	if err != nil {
		return "", fmt.Errorf("could not write to kubeconfig file")
	}

	logsd := []byte("logs")
	err = ioutil.WriteFile(path.Join(assetsDir, "foo", "kismatic.log"), logsd, 0644)
	if err != nil {
		return "", fmt.Errorf("could not write to kismatic.log file")
	}

	return assetsDir, nil
}