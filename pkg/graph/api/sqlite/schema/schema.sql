-- Enable foreign keys
PRAGMA foreign_keys=ON;

-- Create graphs table
CREATE TABLE IF NOT EXISTS graphs (
    uid TEXT PRIMARY KEY,
    label TEXT,
    attrs TEXT
);

-- Create nodes table with a foreign key to graphs
CREATE TABLE IF NOT EXISTS nodes (
    uid TEXT PRIMARY KEY,
    id INTEGER UNIQUE NOT NULL,
    graph TEXT NOT NULL,
    label TEXT,
    attrs TEXT,
    FOREIGN KEY (graph) REFERENCES graphs (uid) ON DELETE CASCADE
);

-- Create edges table with foreign keys to nodes and graphs
CREATE TABLE IF NOT EXISTS edges (
    uid TEXT PRIMARY KEY,
    graph TEXT NOT NULL,
    from_node TEXT NOT NULL,
    to_node TEXT NOT NULL,
    label TEXT,
    weight REAL DEFAULT 1.0,
    attrs TEXT,
    FOREIGN KEY (from_node) REFERENCES nodes (uid) ON DELETE CASCADE,
    FOREIGN KEY (to_node) REFERENCES nodes (uid) ON DELETE CASCADE,
    FOREIGN KEY (graph) REFERENCES graphs (uid) ON DELETE CASCADE
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_graphs_label ON graphs (label);
CREATE INDEX IF NOT EXISTS idx_nodes_id ON nodes (id);
CREATE INDEX IF NOT EXISTS idx_nodes_label ON nodes (label);
CREATE INDEX IF NOT EXISTS idx_nodes_graph_uid ON nodes (graph, uid);
CREATE INDEX IF NOT EXISTS idx_edges_from_node ON edges (from_node);
CREATE INDEX IF NOT EXISTS idx_edges_to_node ON edges (to_node);
CREATE INDEX IF NOT EXISTS idx_edges_label ON edges (label);
CREATE INDEX IF NOT EXISTS idx_edges_from_to_node ON edges (from_node, to_node);
CREATE INDEX IF NOT EXISTS idx_edges_graph_from_to_node ON edges (graph, from_node, to_node);
