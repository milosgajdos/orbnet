package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/milosgajdos/orbnet/pkg/graph/api"
	"github.com/milosgajdos/orbnet/pkg/graph/api/memory"
)

func MustNodeService(t *testing.T, db *memory.DB) api.NodeService {
	ns, err := memory.NewNodeService(db)
	if err != nil {
		t.Fatal(err)
	}
	return ns
}

func TestGetNodes(t *testing.T) {
	t.Run("200_NoFilter", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.NodeService = MustNodeService(t, db)

		uid := "cc099040-9dab-4f3d-848e-3046912aa281"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/nodes", uid)

		req := httptest.NewRequest("GET", urlPath, nil)

		resp, err := s.app.Test(req)
		if err != nil {
			t.Fatalf("failed to get response: %v", err)
		}
		defer resp.Body.Close()

		if code := resp.StatusCode; code != http.StatusOK {
			t.Fatalf("expected status code: %d, got: %d", http.StatusOK, code)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		ret := new(NodesResponse)
		if err := json.Unmarshal(body, ret); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}

		exp := 6
		if ret.N != exp {
			t.Errorf("expected graphs: %d, got: %d", exp, ret.N)
		}
	})

	t.Run("200_Filter", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.NodeService = MustNodeService(t, db)

		uid := "cc099040-9dab-4f3d-848e-3046912aa281"
		label := "Repo"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/nodes?label=%s", uid, label)

		req := httptest.NewRequest("GET", urlPath, nil)

		resp, err := s.app.Test(req)
		if err != nil {
			t.Fatalf("failed to get response: %v", err)
		}
		defer resp.Body.Close()

		if code := resp.StatusCode; code != http.StatusOK {
			t.Fatalf("expected status code: %d, got: %d", http.StatusOK, code)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		ret := new(NodesResponse)
		if err := json.Unmarshal(body, ret); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}

		exp := 2
		if ret.N != exp {
			t.Errorf("expected graphs with label %s: %d, got: %d", label, exp, ret.N)
		}
	})

	t.Run("400", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.NodeService = MustNodeService(t, db)

		uid := "dflksdjfdlksf"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/nodes", uid)
		req := httptest.NewRequest("GET", urlPath, nil)

		resp, err := s.app.Test(req)
		if err != nil {
			t.Fatalf("failed to get response: %v", err)
		}
		defer resp.Body.Close()

		if code := resp.StatusCode; code != http.StatusBadRequest {
			t.Fatalf("expected status code: %d, got: %d", http.StatusBadRequest, code)
		}
	})

	t.Run("404", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.NodeService = MustNodeService(t, db)

		uid := "97153afd-c434-4ca0-a35b-7467fcd08df1"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/nodes", uid)
		req := httptest.NewRequest("GET", urlPath, nil)

		resp, err := s.app.Test(req)
		if err != nil {
			t.Fatalf("failed to get response: %v", err)
		}
		defer resp.Body.Close()

		if code := resp.StatusCode; code != http.StatusNotFound {
			t.Fatalf("expected status code: %d, got: %d", http.StatusNotFound, code)
		}
	})

	t.Run("500", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.NodeService = MustNodeService(t, db)

		// we simulate the loss of DB connection like this.
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close DB: %v", err)
		}

		uid := "cc099040-9dab-4f3d-848e-3046912aa281"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/nodes", uid)
		req := httptest.NewRequest("GET", urlPath, nil)

		resp, err := s.app.Test(req)
		if err != nil {
			t.Fatalf("failed to get response: %v", err)
		}
		defer resp.Body.Close()

		if code := resp.StatusCode; code != http.StatusInternalServerError {
			t.Fatalf("expected status code: %d, got: %d", http.StatusInternalServerError, code)
		}
	})
}

func TestGetNodeByUID(t *testing.T) {
	t.Run("200", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.NodeService = MustNodeService(t, db)

		guid := "cc099040-9dab-4f3d-848e-3046912aa281"
		nuid := "014e5c91-5d5e-4d34-8284-4354aa9f62cd"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/nodes/uid/%s", guid, nuid)
		req := httptest.NewRequest("GET", urlPath, nil)

		resp, err := s.app.Test(req)
		if err != nil {
			t.Fatalf("failed to get response: %v", err)
		}
		defer resp.Body.Close()

		if code := resp.StatusCode; code != http.StatusOK {
			t.Fatalf("expected status code: %d, got: %d", http.StatusOK, code)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		n := new(api.Node)
		if err := json.Unmarshal(body, n); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}

		if n.UID != nuid {
			t.Fatalf("expected graph uid: %s, got: %s", nuid, n.UID)
		}
	})

	t.Run("400", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.NodeService = MustNodeService(t, db)

		testCases := []struct {
			guid string
			nuid string
		}{
			{"dflksdjfdlksf", "014e5c91-5d5e-4d34-8284-4354aa9f62cd"},
			{"cc099040-9dab-4f3d-848e-3046912aa281", "randuid"},
		}

		for _, tc := range testCases {
			guid := tc.guid
			nuid := tc.nuid
			urlPath := fmt.Sprintf("/api/v1/graphs/%s/nodes/uid/%s", guid, nuid)
			req := httptest.NewRequest("GET", urlPath, nil)

			resp, err := s.app.Test(req)
			if err != nil {
				t.Fatalf("failed to get response: %v", err)
			}
			defer resp.Body.Close()

			if code := resp.StatusCode; code != http.StatusBadRequest {
				t.Fatalf(" expected status code: %d, got: %d, for uid: %q", http.StatusBadRequest, code, tc.nuid)
			}
		}
	})

	t.Run("404", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.NodeService = MustNodeService(t, db)

		testCases := []struct {
			guid string
			nuid string
		}{
			// Node not found
			{"cc099040-9dab-4f3d-848e-3046912aa281", "97153afd-c434-4ca0-a35b-7467fcd08df1"},
			// graph not found
			{"97153afd-c434-4ca0-a35b-7467fcd08df1", "014e5c91-5d5e-4d34-8284-4354aa9f62cd"},
		}

		for _, tc := range testCases {
			guid := tc.guid
			nuid := tc.nuid
			urlPath := fmt.Sprintf("/api/v1/graphs/%s/nodes/uid/%s", guid, nuid)
			req := httptest.NewRequest("GET", urlPath, nil)

			resp, err := s.app.Test(req)
			if err != nil {
				t.Fatalf("failed to get response: %v", err)
			}
			defer resp.Body.Close()

			if code := resp.StatusCode; code != http.StatusNotFound {
				t.Fatalf("expected status code: %d, got: %d", http.StatusNotFound, code)
			}
		}
	})

	t.Run("500", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.NodeService = MustNodeService(t, db)

		// we simulate the loss of DB connection like this.
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close DB: %v", err)
		}

		guid := "cc099040-9dab-4f3d-848e-3046912aa281"
		nuid := "014e5c91-5d5e-4d34-8284-4354aa9f62cd"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/nodes/uid/%s", guid, nuid)
		req := httptest.NewRequest("GET", urlPath, nil)

		resp, err := s.app.Test(req)
		if err != nil {
			t.Fatalf("failed to get response: %v", err)
		}
		defer resp.Body.Close()

		if code := resp.StatusCode; code != http.StatusInternalServerError {
			t.Fatalf("expected status code: %d, got: %d", http.StatusInternalServerError, code)
		}
	})
}

func TestGetNodeByID(t *testing.T) {
	t.Run("200", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.NodeService = MustNodeService(t, db)

		guid := "cc099040-9dab-4f3d-848e-3046912aa281"
		id := int64(0)
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/nodes/%d", guid, id)
		req := httptest.NewRequest("GET", urlPath, nil)

		resp, err := s.app.Test(req)
		if err != nil {
			t.Fatalf("failed to get response: %v", err)
		}
		defer resp.Body.Close()

		if code := resp.StatusCode; code != http.StatusOK {
			t.Fatalf("expected status code: %d, got: %d", http.StatusOK, code)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		n := new(api.Node)
		if err := json.Unmarshal(body, n); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}

		if n.ID != id {
			t.Fatalf("expected node id: %d, got: %d", id, n.ID)
		}
	})

	t.Run("400", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.NodeService = MustNodeService(t, db)

		testCases := []struct {
			guid string
			nuid string
		}{
			{"sdlfkjsdflkdjf", "0"},
			{"cc099040-9dab-4f3d-848e-3046912aa281", "1.3"},
			{"97153afd-c434-4ca0-a35b-7467fcd08df1", "-10"},
		}

		for _, tc := range testCases {
			guid := tc.guid
			nuid := tc.nuid
			urlPath := fmt.Sprintf("/api/v1/graphs/%s/nodes/%s", guid, nuid)
			req := httptest.NewRequest("GET", urlPath, nil)

			resp, err := s.app.Test(req)
			if err != nil {
				t.Fatalf("failed to get response: %v", err)
			}
			defer resp.Body.Close()

			if code := resp.StatusCode; code != http.StatusBadRequest {
				t.Fatalf(" expected status code: %d, got: %d, for uid: %q", http.StatusBadRequest, code, tc.nuid)
			}
		}
	})

	t.Run("404", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.NodeService = MustNodeService(t, db)

		testCases := []struct {
			guid string
			nuid string
		}{
			{"cc099040-9dab-4f3d-848e-3046912aa281", "123456"},
			{"97153afd-c434-4ca0-a35b-7467fcd08df1", "7890"},
		}

		for _, tc := range testCases {
			guid := tc.guid
			nuid := tc.nuid
			urlPath := fmt.Sprintf("/api/v1/graphs/%s/nodes/%s", guid, nuid)
			req := httptest.NewRequest("GET", urlPath, nil)

			resp, err := s.app.Test(req)
			if err != nil {
				t.Fatalf("failed to get response: %v", err)
			}
			defer resp.Body.Close()

			if code := resp.StatusCode; code != http.StatusNotFound {
				t.Fatalf("expected status code: %d, got: %d", http.StatusNotFound, code)
			}
		}
	})

	t.Run("500", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.NodeService = MustNodeService(t, db)

		// we simulate the loss of DB connection like this.
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close DB: %v", err)
		}

		guid := "97153afd-c434-4ca0-a35b-7467fcd08df1"
		nuid := "0"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/nodes/%s", guid, nuid)
		req := httptest.NewRequest("GET", urlPath, nil)

		resp, err := s.app.Test(req)
		if err != nil {
			t.Fatalf("failed to get response: %v", err)
		}
		defer resp.Body.Close()

		if code := resp.StatusCode; code != http.StatusInternalServerError {
			t.Fatalf("expected status code: %d, got: %d", http.StatusInternalServerError, code)
		}
	})
}

func TestCreateNode(t *testing.T) {
	t.Run("200", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.NodeService = MustNodeService(t, db)

		label := "foolabel"
		attrs := map[string]interface{}{
			"one": 1.0,
			"two": "twostring",
		}
		apiNode := &api.Node{
			Label: StringPtr(label),
			Attrs: attrs,
		}

		testBody, err := json.Marshal(apiNode)
		if err != nil {
			t.Fatalf("failed to serialise req body: %v", err)
		}

		uid := "cc099040-9dab-4f3d-848e-3046912aa281"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/nodes", uid)
		req := httptest.NewRequest("POST", urlPath, bytes.NewReader(testBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.app.Test(req)
		if err != nil {
			t.Fatalf("failed to get response: %v", err)
		}
		defer resp.Body.Close()

		if code := resp.StatusCode; code != http.StatusOK {
			t.Fatalf("expected status code: %d, got: %d", http.StatusOK, code)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		g := new(api.Node)
		if err := json.Unmarshal(body, g); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}

		if *g.Label != label {
			t.Fatalf("expected label: %s, got: %s", label, *g.Label)
		}

		if !reflect.DeepEqual(g.Attrs, attrs) {
			t.Fatalf("expected attrs: %s, got: %s", attrs, g.Attrs)
		}
	})

	t.Run("400", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.NodeService = MustNodeService(t, db)

		label := "foolabel"
		attrs := map[string]interface{}{
			"one": 1,
			"two": "twostring",
		}
		apiNode := &api.Node{
			Label: StringPtr(label),
			Attrs: attrs,
		}

		testBody, err := json.Marshal(apiNode)
		if err != nil {
			t.Fatalf("failed to serialise req body: %v", err)
		}

		uid := "dsdsdsffsfs"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/nodes", uid)
		req := httptest.NewRequest("POST", urlPath, bytes.NewReader(testBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.app.Test(req)
		if err != nil {
			t.Fatalf("failed to get response: %v", err)
		}
		defer resp.Body.Close()

		if code := resp.StatusCode; code != http.StatusBadRequest {
			t.Fatalf("expected status code: %d, got: %d", http.StatusBadRequest, code)
		}
	})

	t.Run("404", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.NodeService = MustNodeService(t, db)

		label := "foolabel"
		attrs := map[string]interface{}{
			"one": 1,
			"two": "twostring",
		}
		apiNode := &api.Node{
			Label: StringPtr(label),
			Attrs: attrs,
		}

		testBody, err := json.Marshal(apiNode)
		if err != nil {
			t.Fatalf("failed to serialise req body: %v", err)
		}

		uid := "97153afd-c434-4ca0-a35b-7467fcd08df1"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/nodes", uid)
		req := httptest.NewRequest("POST", urlPath, bytes.NewReader(testBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.app.Test(req)
		if err != nil {
			t.Fatalf("failed to get response: %v", err)
		}
		defer resp.Body.Close()

		if code := resp.StatusCode; code != http.StatusNotFound {
			t.Fatalf("expected status code: %d, got: %d", http.StatusNotFound, code)
		}
	})

	t.Run("500", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.NodeService = MustNodeService(t, db)

		// we simulate the loss of DB connection like this.
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close DB: %v", err)
		}

		label := "foolabel"
		attrs := map[string]interface{}{
			"one": 1,
			"two": "twostring",
		}
		apiNode := &api.Node{
			Label: StringPtr(label),
			Attrs: attrs,
		}

		testBody, err := json.Marshal(apiNode)
		if err != nil {
			t.Fatalf("failed to serialise req body: %v", err)
		}

		uid := "cc099040-9dab-4f3d-848e-3046912aa281"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/nodes", uid)
		req := httptest.NewRequest("POST", urlPath, bytes.NewReader(testBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.app.Test(req)
		if err != nil {
			t.Fatalf("failed to get response: %v", err)
		}
		defer resp.Body.Close()

		if code := resp.StatusCode; code != http.StatusInternalServerError {
			t.Fatalf("expected status code: %d, got: %d", http.StatusInternalServerError, code)
		}
	})
}

func TestUpdateNode(t *testing.T) {
	t.Run("200", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.NodeService = MustNodeService(t, db)

		label := "foolabel"
		update := &api.NodeUpdate{
			Label: &label,
		}

		testBody, err := json.Marshal(update)
		if err != nil {
			t.Fatalf("failed to serialise req body: %v", err)
		}

		guid := "cc099040-9dab-4f3d-848e-3046912aa281"
		id := "0"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/nodes/%s", guid, id)
		req := httptest.NewRequest("PATCH", urlPath, bytes.NewReader(testBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.app.Test(req)
		if err != nil {
			t.Fatalf("failed to get response: %v", err)
		}
		defer resp.Body.Close()

		if code := resp.StatusCode; code != http.StatusOK {
			t.Fatalf("expected status code: %d, got: %d", http.StatusOK, code)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		g := new(api.Node)
		if err := json.Unmarshal(body, g); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}

		if *g.Label != label {
			t.Fatalf("expected label: %s, got: %s", label, *g.Label)
		}
	})

	t.Run("400", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.NodeService = MustNodeService(t, db)

		testCases := []struct {
			guid string
			id   string
		}{
			{"sdlfkjsdflkdjf", "0"},
			{"cc099040-9dab-4f3d-848e-3046912aa281", "1.3"},
			{"97153afd-c434-4ca0-a35b-7467fcd08df1", "-10"},
		}

		for _, tc := range testCases {
			urlPath := fmt.Sprintf("/api/v1/graphs/%s/nodes/%s", tc.guid, tc.id)
			req := httptest.NewRequest("PATCH", urlPath, nil)
			req.Header.Set("Content-Type", "application/json")

			resp, err := s.app.Test(req)
			if err != nil {
				t.Fatalf("failed to get response: %v", err)
			}
			defer resp.Body.Close()

			if code := resp.StatusCode; code != http.StatusBadRequest {
				t.Fatalf("expected status code: %d, got: %d", http.StatusBadRequest, code)
			}
		}
	})

	t.Run("404", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.NodeService = MustNodeService(t, db)

		label := "foolabel"
		update := &api.NodeUpdate{
			Label: &label,
		}

		testBody, err := json.Marshal(update)
		if err != nil {
			t.Fatalf("failed to serialise req body: %v", err)
		}

		// this graph does not exist
		guid := "97153afd-c434-4ca0-a35b-7467fcd08df1"
		id := "0"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/nodes/%s", guid, id)
		req := httptest.NewRequest("PATCH", urlPath, bytes.NewReader(testBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.app.Test(req)
		if err != nil {
			t.Fatalf("failed to get response: %v", err)
		}
		defer resp.Body.Close()

		if code := resp.StatusCode; code != http.StatusNotFound {
			t.Fatalf("expected status code: %d, got: %d", http.StatusNotFound, code)
		}
	})

	t.Run("500", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.NodeService = MustNodeService(t, db)

		// we simulate the loss of DB connection like this.
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close DB: %v", err)
		}

		label := "foolabel"
		update := &api.NodeUpdate{
			Label: &label,
		}

		testBody, err := json.Marshal(update)
		if err != nil {
			t.Fatalf("failed to serialise req body: %v", err)
		}

		guid := "cc099040-9dab-4f3d-848e-3046912aa281"
		id := "0"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/nodes/%s", guid, id)
		req := httptest.NewRequest("PATCH", urlPath, bytes.NewReader(testBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.app.Test(req)
		if err != nil {
			t.Fatalf("failed to get response: %v", err)
		}
		defer resp.Body.Close()

		if code := resp.StatusCode; code != http.StatusInternalServerError {
			t.Fatalf("expected status code: %d, got: %d", http.StatusInternalServerError, code)
		}
	})
}

func TestDeleteNodeByID(t *testing.T) {
	t.Run("204", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.NodeService = MustNodeService(t, db)

		uid := "cc099040-9dab-4f3d-848e-3046912aa281"
		id := int64(0)
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/nodes/%d", uid, id)
		req := httptest.NewRequest("DELETE", urlPath, nil)

		resp, err := s.app.Test(req)
		if err != nil {
			t.Fatalf("failed to get response: %v", err)
		}
		defer resp.Body.Close()

		if code := resp.StatusCode; code != http.StatusNoContent {
			t.Fatalf("expected status code: %d, got: %d", http.StatusNoContent, code)
		}
	})

	t.Run("400", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.NodeService = MustNodeService(t, db)

		testCases := []struct {
			guid string
			id   string
		}{
			{"sdlfkjsdflkdjf", "0"},
			{"cc099040-9dab-4f3d-848e-3046912aa281", "1.3"},
			{"97153afd-c434-4ca0-a35b-7467fcd08df1", "-10"},
		}

		for _, tc := range testCases {
			urlPath := fmt.Sprintf("/api/v1/graphs/%s/nodes/%s", tc.guid, tc.id)
			req := httptest.NewRequest("DELETE", urlPath, nil)

			resp, err := s.app.Test(req)
			if err != nil {
				t.Fatalf("failed to get response: %v", err)
			}
			defer resp.Body.Close()

			if code := resp.StatusCode; code != http.StatusBadRequest {
				t.Fatalf("expected status code: %d, got: %d", http.StatusBadRequest, code)
			}
		}
	})

	t.Run("404", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.NodeService = MustNodeService(t, db)

		guid := "97153afd-c434-4ca0-a35b-7467fcd08df1"
		id := "0"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/nodes/%s", guid, id)
		req := httptest.NewRequest("DELETE", urlPath, nil)

		resp, err := s.app.Test(req)
		if err != nil {
			t.Fatalf("failed to get response: %v", err)
		}
		defer resp.Body.Close()

		if code := resp.StatusCode; code != http.StatusNotFound {
			t.Fatalf("expected status code: %d, got: %d", http.StatusNotFound, code)
		}
	})

	t.Run("500", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.NodeService = MustNodeService(t, db)

		// we simulate the loss of DB connection like this.
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close DB: %v", err)
		}

		uid := "97153afd-c434-4ca0-a35b-7467fcd08df1"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/nodes/%d", uid, 0)
		req := httptest.NewRequest("DELETE", urlPath, nil)

		resp, err := s.app.Test(req)
		if err != nil {
			t.Fatalf("failed to get response: %v", err)
		}
		defer resp.Body.Close()

		if code := resp.StatusCode; code != http.StatusInternalServerError {
			t.Fatalf("expected status code: %d, got: %d", http.StatusInternalServerError, code)
		}
	})
}

func TestDeleteNodeByUID(t *testing.T) {
	t.Run("204", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.NodeService = MustNodeService(t, db)

		guid := "cc099040-9dab-4f3d-848e-3046912aa281"
		uid := "014e5c91-5d5e-4d34-8284-4354aa9f62cd"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/nodes/uid/%s", guid, uid)
		req := httptest.NewRequest("DELETE", urlPath, nil)

		resp, err := s.app.Test(req)
		if err != nil {
			t.Fatalf("failed to get response: %v", err)
		}
		defer resp.Body.Close()

		if code := resp.StatusCode; code != http.StatusNoContent {
			t.Fatalf("expected status code: %d, got: %d", http.StatusNoContent, code)
		}
	})

	t.Run("400", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.NodeService = MustNodeService(t, db)

		testCases := []struct {
			guid string
			uid  string
		}{
			{"sdlfkjsdflkdjf", "014e5c91-5d5e-4d34-8284-4354aa9f62cd"},
			{"cc099040-9dab-4f3d-848e-3046912aa281", "dfsdfdfdf"},
		}

		for _, tc := range testCases {
			urlPath := fmt.Sprintf("/api/v1/graphs/%s/nodes/uid/%s", tc.guid, tc.uid)
			req := httptest.NewRequest("DELETE", urlPath, nil)

			resp, err := s.app.Test(req)
			if err != nil {
				t.Fatalf("failed to get response: %v", err)
			}
			defer resp.Body.Close()

			if code := resp.StatusCode; code != http.StatusBadRequest {
				t.Fatalf("expected status code: %d, got: %d", http.StatusBadRequest, code)
			}
		}
	})

	t.Run("404", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.NodeService = MustNodeService(t, db)

		guid := "97153afd-c434-4ca0-a35b-7467fcd08df1"
		id := "014e5c91-5d5e-4d34-8284-4354aa9f62cd"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/nodes/uid/%s", guid, id)
		req := httptest.NewRequest("DELETE", urlPath, nil)

		resp, err := s.app.Test(req)
		if err != nil {
			t.Fatalf("failed to get response: %v", err)
		}
		defer resp.Body.Close()

		if code := resp.StatusCode; code != http.StatusNotFound {
			t.Fatalf("expected status code: %d, got: %d", http.StatusNotFound, code)
		}
	})

	t.Run("500", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.NodeService = MustNodeService(t, db)

		// we simulate the loss of DB connection like this.
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close DB: %v", err)
		}

		guid := "97153afd-c434-4ca0-a35b-7467fcd08df1"
		uid := "014e5c91-5d5e-4d34-8284-4354aa9f62cd"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/nodes/uid/%s", guid, uid)
		req := httptest.NewRequest("DELETE", urlPath, nil)

		resp, err := s.app.Test(req)
		if err != nil {
			t.Fatalf("failed to get response: %v", err)
		}
		defer resp.Body.Close()

		if code := resp.StatusCode; code != http.StatusInternalServerError {
			t.Fatalf("expected status code: %d, got: %d", http.StatusInternalServerError, code)
		}
	})
}
