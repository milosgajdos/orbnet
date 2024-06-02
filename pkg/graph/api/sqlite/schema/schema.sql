-- Enable foreign keys
PRAGMA foreign_keys=ON;

-- Create graphs table
CREATE TABLE IF NOT EXISTS graphs (
    uid TEXT PRIMARY KEY NOT NULL CHECK(uid <> ''),
    label TEXT,
    attrs TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- Create nodes table with a foreign key to graphs
CREATE TABLE IF NOT EXISTS nodes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid TEXT UNIQUE NOT NULL CHECK(uid <> ''),
    graph TEXT NOT NULL,
    label TEXT,
    attrs TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (graph) REFERENCES graphs (uid) ON DELETE CASCADE
);

-- Create edges table with foreign keys to nodes and graphs
CREATE TABLE IF NOT EXISTS edges (
    uid TEXT PRIMARY KEY NOT NULL CHECK(uid <> ''),
    graph TEXT NOT NULL,
    source TEXT NOT NULL,
    target TEXT NOT NULL,
    label TEXT,
    weight REAL DEFAULT 1.0,
    attrs TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (source) REFERENCES nodes (uid) ON DELETE CASCADE,
    FOREIGN KEY (target) REFERENCES nodes (uid) ON DELETE CASCADE,
    FOREIGN KEY (graph) REFERENCES graphs (uid) ON DELETE CASCADE
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_graphs_label ON graphs (label);
CREATE INDEX IF NOT EXISTS idx_nodes_label ON nodes (label);
CREATE INDEX IF NOT EXISTS idx_nodes_graph_uid ON nodes (graph, uid);
CREATE INDEX IF NOT EXISTS idx_nodes_graph_id ON nodes (graph, id);
CREATE INDEX IF NOT EXISTS idx_edges_source ON edges (source);
CREATE INDEX IF NOT EXISTS idx_edges_target ON edges (target);
CREATE INDEX IF NOT EXISTS idx_edges_label ON edges (label);
CREATE INDEX IF NOT EXISTS idx_edges_from_target ON edges (source, target);
CREATE INDEX IF NOT EXISTS idx_edges_graph_from_target ON edges (graph, source, target);
