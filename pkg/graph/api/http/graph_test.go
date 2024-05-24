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

func MustGraphService(t *testing.T, db *memory.DB) api.GraphService {
	gs, err := memory.NewGraphService(db)
	if err != nil {
		t.Fatal(err)
	}
	return gs
}

func TestGetAllGraphs(t *testing.T) {
	t.Run("200", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		gs := MustGraphService(t, db)
		s.GraphService = gs

		req := httptest.NewRequest("GET", "/api/v1/graphs", nil)

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

		ret := new(GraphsResponse)
		if err := json.Unmarshal(body, ret); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}

		exp := 2
		if ret.N != exp {
			t.Errorf("expected graphs: %d, got: %d", exp, ret.N)
		}
	})

	t.Run("500", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		gs := MustGraphService(t, db)
		s.GraphService = gs

		// we simulate the loss of DB connection like this.
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close DB: %v", err)
		}

		req := httptest.NewRequest("GET", "/api/v1/graphs", nil)

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

func TestGraphByUID(t *testing.T) {
	t.Run("200", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		gs := MustGraphService(t, db)
		s.GraphService = gs

		uid := "cc099040-9dab-4f3d-848e-3046912aa281"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s", uid)
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

		g := new(api.Graph)
		if err := json.Unmarshal(body, g); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}

		if g.UID != uid {
			t.Fatalf("expected graph uid: %s, got: %s", uid, g.UID)
		}
	})

	t.Run("400", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		gs := MustGraphService(t, db)
		s.GraphService = gs

		uid := "dflksdjfdlksf"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s", uid)
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
		gs := MustGraphService(t, db)
		s.GraphService = gs

		uid := "97153afd-c434-4ca0-a35b-7467fcd08df1"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s", uid)
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
		gs := MustGraphService(t, db)
		s.GraphService = gs

		// we simulate the loss of DB connection like this.
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close DB: %v", err)
		}

		uid := "97153afd-c434-4ca0-a35b-7467fcd08df1"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s", uid)
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

func TestCreateGraph(t *testing.T) {
	t.Run("200", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		gs := MustGraphService(t, db)
		s.GraphService = gs

		label := "foolabel"
		attrs := map[string]interface{}{
			"one": 1.0,
			"two": "twostring",
		}
		apiGraph := &api.Graph{
			Label: &label,
			Attrs: attrs,
		}

		testBody, err := json.Marshal(apiGraph)
		if err != nil {
			t.Fatalf("failed to serialise req body: %v", err)
		}

		req := httptest.NewRequest("POST", "/api/v1/graphs", bytes.NewReader(testBody))
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

		g := new(api.Graph)
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
		gs := MustGraphService(t, db)
		s.GraphService = gs

		label := "foolabel"
		attrs := map[string]interface{}{
			"one": 1,
			"two": "twostring",
		}
		apiGraph := &api.Graph{
			Label: &label,
			Attrs: attrs,
		}

		testBody, err := json.Marshal(apiGraph)
		if err != nil {
			t.Fatalf("failed to serialise req body: %v", err)
		}

		req := httptest.NewRequest("POST", "/api/v1/graphs", bytes.NewReader(testBody))

		resp, err := s.app.Test(req)
		if err != nil {
			t.Fatalf("failed to get response: %v", err)
		}
		defer resp.Body.Close()

		if code := resp.StatusCode; code != http.StatusBadRequest {
			t.Fatalf("expected status code: %d, got: %d", http.StatusBadRequest, code)
		}
	})

	t.Run("500", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		gs := MustGraphService(t, db)
		s.GraphService = gs

		// we simulate the loss of DB connection like this.
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close DB: %v", err)
		}

		label := "foolabel"
		attrs := map[string]interface{}{
			"one": 1,
			"two": "twostring",
		}
		apiGraph := &api.Graph{
			Label: &label,
			Attrs: attrs,
		}

		testBody, err := json.Marshal(apiGraph)
		if err != nil {
			t.Fatalf("failed to serialise req body: %v", err)
		}

		req := httptest.NewRequest("POST", "/api/v1/graphs", bytes.NewReader(testBody))
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

func TestUpdateGraph(t *testing.T) {
	t.Run("200", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		gs := MustGraphService(t, db)
		s.GraphService = gs

		label := "foolabel"
		update := &api.GraphUpdate{
			Label: &label,
		}

		testBody, err := json.Marshal(update)
		if err != nil {
			t.Fatalf("failed to serialise req body: %v", err)
		}

		uid := "cc099040-9dab-4f3d-848e-3046912aa281"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s", uid)
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

		g := new(api.Graph)
		if err := json.Unmarshal(body, g); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}

		if *g.Label != label {
			t.Fatalf("expected label: %s, got: %s", label, *g.Label)
		}
	})

	t.Run("404", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		gs := MustGraphService(t, db)
		s.GraphService = gs

		label := "foolabel"
		update := &api.GraphUpdate{
			Label: &label,
		}

		testBody, err := json.Marshal(update)
		if err != nil {
			t.Fatalf("failed to serialise req body: %v", err)
		}

		uid := "97153afd-c434-4ca0-a35b-7467fcd08df1"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s", uid)
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

	t.Run("400", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		gs := MustGraphService(t, db)
		s.GraphService = gs

		// unparsable UID should 400
		uid := "sdlfkjsdflkdjf"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s", uid)
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
	})

	t.Run("500", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		gs := MustGraphService(t, db)
		s.GraphService = gs

		// we simulate the loss of DB connection like this.
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close DB: %v", err)
		}

		label := "foolabel"
		update := &api.GraphUpdate{
			Label: &label,
		}

		testBody, err := json.Marshal(update)
		if err != nil {
			t.Fatalf("failed to serialise req body: %v", err)
		}

		uid := "cc099040-9dab-4f3d-848e-3046912aa281"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s", uid)
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

func TestDeleteGraph(t *testing.T) {
	t.Run("204", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		gs := MustGraphService(t, db)
		s.GraphService = gs

		uid := "97153afd-c434-4ca0-a35b-7467fcd08df1"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s", uid)
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
		gs := MustGraphService(t, db)
		s.GraphService = gs

		uid := "sdlfkjsdflkdjf"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s", uid)
		req := httptest.NewRequest("DELETE", urlPath, nil)

		resp, err := s.app.Test(req)
		if err != nil {
			t.Fatalf("failed to get response: %v", err)
		}
		defer resp.Body.Close()

		if code := resp.StatusCode; code != http.StatusBadRequest {
			t.Fatalf("expected status code: %d, got: %d", http.StatusBadRequest, code)
		}
	})

	t.Run("500", func(t *testing.T) {
		s := MustServer(t)
		db := MustOpenDB(t, testDir)
		s.GraphService = MustGraphService(t, db)

		// we simulate the loss of DB connection like this.
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close DB: %v", err)
		}

		uid := "97153afd-c434-4ca0-a35b-7467fcd08df1"
		urlPath := fmt.Sprintf("/api/v1/graphs/%s", uid)
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
