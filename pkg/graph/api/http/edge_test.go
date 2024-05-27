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

func MustEdgeService(t *testing.T, db *memory.DB) api.EdgeService {
	es, err := memory.NewEdgeService(db)
	if err != nil {
		t.Fatal(err)
	}
	return es
}

func TestGetAllEdges(t *testing.T) {
	t.Run("200_NoFilter", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.EdgeService = MustEdgeService(t, db)

		uid := "cc099040-9dab-4f3d-848e-3046912aa281"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/edges", uid)

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

		ret := new(EdgesResponse)
		if err := json.Unmarshal(body, ret); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}

		exp := 6
		if ret.N != exp {
			t.Errorf("expected edges: %d, got: %d", exp, ret.N)
		}
	})

	t.Run("200_Filter", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.EdgeService = MustEdgeService(t, db)

		uid := "cc099040-9dab-4f3d-848e-3046912aa281"
		label := "HasOwner"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/edges?label=%s", uid, label)

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

		ret := new(EdgesResponse)
		if err := json.Unmarshal(body, ret); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}

		exp := 4
		if ret.N != exp {
			t.Errorf("expected edges: %d, got: %d", exp, ret.N)
		}
	})

	t.Run("404", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.EdgeService = MustEdgeService(t, db)

		uid := "97153afd-c434-4ca0-a35b-7467fcd08df1"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/edges", uid)
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
		s.EdgeService = MustEdgeService(t, db)

		// we simulate the loss of DB connection like this.
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close DB: %v", err)
		}

		uid := "cc099040-9dab-4f3d-848e-3046912aa281"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/edges", uid)
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

func TestGetEdgeByUID(t *testing.T) {
	t.Run("200", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.EdgeService = MustEdgeService(t, db)

		guid := "cc099040-9dab-4f3d-848e-3046912aa281"
		euid := "b2f42acb-fea2-4584-902b-0c4a01b51037"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/edges/%s", guid, euid)
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

		e := new(api.Edge)
		if err := json.Unmarshal(body, e); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}

		if e.UID != euid {
			t.Fatalf("expected edge uid: %s, got: %s", euid, e.UID)
		}
	})

	t.Run("404", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.EdgeService = MustEdgeService(t, db)

		testCases := []struct {
			guid string
			euid string
		}{
			// graph not found
			{"97153afd-c434-4ca0-a35b-7467fcd08df1", "b2f42acb-fea2-4584-902b-0c4a01b51037"},
			// edge not found
			{"cc099040-9dab-4f3d-848e-3046912aa281", "97153afd-c434-4ca0-a35b-7467fcd08df1"},
		}

		for _, tc := range testCases {
			guid := tc.guid
			euid := tc.euid
			urlPath := fmt.Sprintf("/api/v1/graphs/%s/edges/%s", guid, euid)
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
		s.EdgeService = MustEdgeService(t, db)

		// we simulate the loss of DB connection like this.
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close DB: %v", err)
		}

		guid := "97153afd-c434-4ca0-a35b-7467fcd08df1"
		euid := "b2f42acb-fea2-4584-902b-0c4a01b51037"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/edges/%s", guid, euid)
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

func TestCreateEdge(t *testing.T) {
	t.Run("200", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.EdgeService = MustEdgeService(t, db)

		label := "foolabel"
		attrs := map[string]interface{}{
			"one": 1.0,
			"two": "twostring",
		}
		apiEdge := &api.Edge{
			Source: "a877d937-673d-4b43-bb5c-8f9387e43298",
			Target: "3ba4972d-c780-4308-9ca8-fe466b60da20",
			Label:  label,
			Attrs:  attrs,
		}

		testBody, err := json.Marshal(apiEdge)
		if err != nil {
			t.Fatalf("failed to serialise req body: %v", err)
		}

		uid := "cc099040-9dab-4f3d-848e-3046912aa281"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/edges", uid)
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

		e := new(api.Edge)
		if err := json.Unmarshal(body, e); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}

		if e.Label != label {
			t.Fatalf("expected label: %s, got: %s", label, e.Label)
		}

		if !reflect.DeepEqual(e.Attrs, attrs) {
			t.Fatalf("expected attrs: %s, got: %s", attrs, e.Attrs)
		}
	})

	t.Run("400", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.EdgeService = MustEdgeService(t, db)

		label := "foolabel"
		attrs := map[string]interface{}{
			"one": 1,
			"two": "twostring",
		}
		apiEdge := &api.Edge{
			Source: "a877d937-673d-4b43-bb5c-8f9387e43298",
			Target: "a877d937-673d-4b43-bb5c-8f9387e43298",
			Label:  label,
			Attrs:  attrs,
		}

		testBody, err := json.Marshal(apiEdge)
		if err != nil {
			t.Fatalf("failed to serialise req body: %v", err)
		}

		guid := "sdfdfdfd"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/edges", guid)
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
		s.EdgeService = MustEdgeService(t, db)

		label := "foolabel"
		attrs := map[string]interface{}{
			"one": 1,
			"two": "twostring",
		}
		apiEdge := &api.Edge{
			Source: "foo",
			Target: "bar",
			Label:  label,
			Attrs:  attrs,
		}

		testBody, err := json.Marshal(apiEdge)
		if err != nil {
			t.Fatalf("failed to serialise req body: %v", err)
		}

		uid := "97153afd-c434-4ca0-a35b-7467fcd08df1"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/edges", uid)
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
		s.EdgeService = MustEdgeService(t, db)

		// we simulate the loss of DB connection like this.
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close DB: %v", err)
		}

		label := "foolabel"
		attrs := map[string]interface{}{
			"one": 1,
			"two": "twostring",
		}
		apiEdge := &api.Edge{
			Source: "foo",
			Target: "bar",
			Label:  label,
			Attrs:  attrs,
		}

		testBody, err := json.Marshal(apiEdge)
		if err != nil {
			t.Fatalf("failed to serialise req body: %v", err)
		}

		uid := "cc099040-9dab-4f3d-848e-3046912aa281"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/edges", uid)
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

func TestDeleteEdgeByUID(t *testing.T) {
	t.Run("200", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.EdgeService = MustEdgeService(t, db)

		guid := "cc099040-9dab-4f3d-848e-3046912aa281"
		euid := "b2f42acb-fea2-4584-902b-0c4a01b51037"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/edges/%s", guid, euid)
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

	t.Run("404", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.EdgeService = MustEdgeService(t, db)

		testCases := []struct {
			guid string
			euid string
		}{
			// graph not found
			{"97153afd-c434-4ca0-a35b-7467fcd08df1", "b2f42acb-fea2-4584-902b-0c4a01b51037"},
			// edge not found
			{"cc099040-9dab-4f3d-848e-3046912aa281", "97153afd-c434-4ca0-a35b-7467fcd08df1"},
		}

		for _, tc := range testCases {
			guid := tc.guid
			euid := tc.euid
			urlPath := fmt.Sprintf("/api/v1/graphs/%s/edges/%s", guid, euid)
			req := httptest.NewRequest("DELETE", urlPath, nil)

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
		s.EdgeService = MustEdgeService(t, db)

		// we simulate the loss of DB connection like this.
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close DB: %v", err)
		}

		guid := "97153afd-c434-4ca0-a35b-7467fcd08df1"
		euid := "b2f42acb-fea2-4584-902b-0c4a01b51037"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/edges/%s", guid, euid)
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

func TestUpdateEdgeBetween(t *testing.T) {
	t.Run("200", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.EdgeService = MustEdgeService(t, db)

		label := "foolabel"
		update := &api.NodeUpdate{
			Label: &label,
		}

		testBody, err := json.Marshal(update)
		if err != nil {
			t.Fatalf("failed to serialise req body: %v", err)
		}

		guid := "cc099040-9dab-4f3d-848e-3046912aa281"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/edges?source=%s&target=%s",
			guid, "a877d937-673d-4b43-bb5c-8f9387e43298", "3ba4972d-c780-4308-9ca8-fe466b60da20")
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
	})

	t.Run("400", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.EdgeService = MustEdgeService(t, db)

		testCases := []struct {
			guid   string
			source string
			target string
		}{
			{"cc099040-9dab-4f3d-848e-3046912aa281", "", "1"},
			{"97153afd-c434-4ca0-a35b-7467fcd08df1", "0", ""},
		}

		for _, tc := range testCases {
			urlPath := fmt.Sprintf("/api/v1/graphs/%s/edges?source=%s&target=%s", tc.guid, tc.source, tc.target)
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
		s.EdgeService = MustEdgeService(t, db)

		label := "foolabel"
		update := &api.NodeUpdate{
			Label: &label,
		}

		testBody, err := json.Marshal(update)
		if err != nil {
			t.Fatalf("failed to serialise req body: %v", err)
		}

		uid := "97153afd-c434-4ca0-a35b-7467fcd08df1"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/edges?source=%d&target=%d", uid, 0, 1)
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
		s.EdgeService = MustEdgeService(t, db)

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

		uid := "97153afd-c434-4ca0-a35b-7467fcd08df1"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/edges?source=%d&target=%d", uid, 0, 1)
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

func TestDeleteEdgeBetween(t *testing.T) {
	t.Run("204", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.EdgeService = MustEdgeService(t, db)

		guid := "cc099040-9dab-4f3d-848e-3046912aa281"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/edges?source=%s&target=%s",
			guid, "a877d937-673d-4b43-bb5c-8f9387e43298", "3ba4972d-c780-4308-9ca8-fe466b60da20")
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
		s.EdgeService = MustEdgeService(t, db)

		testCases := []struct {
			guid   string
			source string
			target string
		}{
			{"cc099040-9dab-4f3d-848e-3046912aa281", "", "1"},
			{"97153afd-c434-4ca0-a35b-7467fcd08df1", "0", ""},
		}

		for _, tc := range testCases {
			urlPath := fmt.Sprintf("/api/v1/graphs/%s/edges?source=%s&target=%s", tc.guid, tc.source, tc.target)
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
		s.EdgeService = MustEdgeService(t, db)

		uid := "97153afd-c434-4ca0-a35b-7467fcd08df1"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/edges?source=%s&target=%s", uid, "foo", "bar")
		req := httptest.NewRequest("DELETE", urlPath, nil)

		resp, err := s.app.Test(req)
		if err != nil {
			t.Fatalf("failed to get response: %v", err)
		}
		defer resp.Body.Close()

		if code := resp.StatusCode; code != http.StatusNotFound {
			t.Fatalf("expected status code: %d, got: %d", http.StatusNoContent, code)
		}
	})

	t.Run("500", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.EdgeService = MustEdgeService(t, db)

		// we simulate the loss of DB connection like this.
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close DB: %v", err)
		}

		uid := "97153afd-c434-4ca0-a35b-7467fcd08df1"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s/edges?source=%s&target=%s", uid, "foo", "bar")
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
