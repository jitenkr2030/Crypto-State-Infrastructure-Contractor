// CSIC Platform - Blockchain Node Manager
// Manages Bitcoin and Ethereum node connections

package node

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/csic-platform/blockchain/nodes/bitcoin"
    "github.com/csic-platform/blockchain/nodes/ethereum"
)

// NodeManager manages blockchain node connections
type NodeManager struct {
    bitcoinNode  *bitcoin.Node
    ethereumNode *ethereum.Node
    mu           sync.RWMutex
    status       map[string]NodeStatus
}

// NodeStatus represents the status of a blockchain node
type NodeStatus struct {
    NodeType   string    `json:"node_type"`
    Connected  bool      `json:"connected"`
    BlockHeight int64    `json:"block_height"`
    LastSync   time.Time `json:"last_sync"`
    Peers      int       `json:"peers"`
    Latency    int64     `json:"latency_ms"`
    Error      string    `json:"error,omitempty"`
}

// BlockInfo represents blockchain block information
type BlockInfo struct {
    Hash          string        `json:"hash"`
    Height        int64         `json:"height"`
    Timestamp     time.Time     `json:"timestamp"`
    Transactions  int           `json:"transaction_count"`
    Size          int64         `json:"size"`
    Confirmations int           `json:"confirmations"`
    PreviousHash  string        `json:"previous_hash"`
    MerkleRoot    string        `json:"merkle_root"`
    Nonce         string        `json:"nonce"`
    Difficulty    float64       `json:"difficulty"`
    FeeRate       float64       `json:"fee_rate"`
}

// TransactionInfo represents blockchain transaction information
type TransactionInfo struct {
    TxID         string        `json:"tx_id"`
    BlockHash    string        `json:"block_hash,omitempty"`
    BlockHeight  int64         `json:"block_height,omitempty"`
    Timestamp    time.Time     `json:"timestamp"`
    Confirmations int          `json:"confirmations"`
    Size         int           `json:"size"`
    Fee          float64       `json:"fee"`
    FeeRate      float64       `json:"fee_rate"`
    Inputs       []TxInput     `json:"inputs"`
    Outputs      []TxOutput    `json:"outputs"`
    Status       string        `json:"status"`
}

// TxInput represents a transaction input
type TxInput struct {
    PreviousTxID string  `json:"previous_tx_id"`
    Index        int     `json:"index"`
    Address      string  `json:"address"`
    Value        float64 `json:"value"`
    ScriptSig    string  `json:"script_sig"`
}

// TxOutput represents a transaction output
type TxOutput struct {
    Index     int     `json:"index"`
    Address   string  `json:"address"`
    Value     float64 `json:"value"`
    ScriptPubKey string `json:"script_pub_key"`
    Spent     bool    `json:"spent"`
}

// AddressInfo represents address information
type AddressInfo struct {
    Address        string         `json:"address"`
    Balance        float64        `json:"balance"`
    Received       float64        `json:"total_received"`
    Sent           float64        `json:"total_sent"`
    UnconfirmedBalance float64    `json:"unconfirmed_balance"`
    TxCount        int            `json:"transaction_count"`
    UTXOs          []UTXO         `json:"utxos"`
}

// UTXO represents an unspent transaction output
type UTXO struct {
    TxID        string    `json:"tx_id"`
    Index       int       `json:"index"`
    Value       float64   `json:"value"`
    ScriptPubKey string   `json:"script_pub_key"`
    Confirmations int     `json:"confirmations"`
    Timestamp   time.Time `json:"timestamp"`
}

// NewNodeManager creates a new node manager
func NewNodeManager(btcConfig bitcoin.Config, ethConfig ethereum.Config) (*NodeManager, error) {
    manager := &NodeManager{
        status: make(map[string]NodeStatus),
    }

    // Initialize Bitcoin node
    btcNode, err := bitcoin.NewNode(btcConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize Bitcoin node: %w", err)
    }
    manager.bitcoinNode = btcNode

    // Initialize Ethereum node
    ethNode, err := ethereum.NewNode(ethConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize Ethereum node: %w", err)
    }
    manager.ethereumNode = ethNode

    return manager, nil
}

// Start starts all blockchain nodes
func (m *NodeManager) Start(ctx context.Context) error {
    var wg sync.WaitGroup
    errChan := make(chan error, 2)

    // Start Bitcoin node
    wg.Add(1)
    go func() {
        defer wg.Done()
        if err := m.bitcoinNode.Start(ctx); err != nil {
            errChan <- fmt.Errorf("Bitcoin node error: %w", err)
        }
    }()

    // Start Ethereum node
    wg.Add(1)
    go func() {
        defer wg.Done()
        if err := m.ethereumNode.Start(ctx); err != nil {
            errChan <- fmt.Errorf("Ethereum node error: %w", err)
        }
    }()

    // Wait for all nodes to start
    wg.Wait()
    close(errChan)

    // Check for errors
    for err := range errChan {
        log.Printf("Node startup error: %v", err)
    }

    // Start status monitoring
    go m.monitorStatus(ctx)

    return nil
}

// Stop stops all blockchain nodes
func (m *NodeManager) Stop() error {
    if err := m.bitcoinNode.Stop(); err != nil {
        return fmt.Errorf("failed to stop Bitcoin node: %w", err)
    }
    if err := m.ethereumNode.Stop(); err != nil {
        return fmt.Errorf("failed to stop Ethereum node: %w", err)
    }
    return nil
}

// GetStatus returns the status of all nodes
func (m *NodeManager) GetStatus() map[string]NodeStatus {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.status
}

// GetBitcoinBlock gets a Bitcoin block by height or hash
func (m *NodeManager) GetBitcoinBlock(ctx context.Context, query string) (*BlockInfo, error) {
    return m.bitcoinNode.GetBlock(ctx, query)
}

// GetEthereumBlock gets an Ethereum block by height or hash
func (m *NodeManager) GetEthereumBlock(ctx context.Context, query string) (*BlockInfo, error) {
    return m.ethereumNode.GetBlock(ctx, query)
}

// GetBitcoinTransaction gets a Bitcoin transaction by txid
func (m *NodeManager) GetBitcoinTransaction(ctx context.Context, txid string) (*TransactionInfo, error) {
    return m.bitcoinNode.GetTransaction(ctx, txid)
}

// GetEthereumTransaction gets an Ethereum transaction by txid
func (m *NodeManager) GetEthereumTransaction(ctx context.Context, txid string) (*TransactionInfo, error) {
    return m.ethereumNode.GetTransaction(ctx, txid)
}

// GetBitcoinAddress gets information about a Bitcoin address
func (m *NodeManager) GetBitcoinAddress(ctx context.Context, address string) (*AddressInfo, error) {
    return m.bitcoinNode.GetAddressInfo(ctx, address)
}

// GetEthereumAddress gets information about an Ethereum address
func (m *NodeManager) GetEthereumAddress(ctx context.Context, address string) (*AddressInfo, error) {
    return m.ethereumNode.GetAddressInfo(ctx, address)
}

// monitorStatus continuously monitors node status
func (m *NodeManager) monitorStatus(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            m.updateStatus()
        }
    }
}

// updateStatus updates the status of all nodes
func (m *NodeManager) updateStatus() {
    m.mu.Lock()
    defer m.mu.Unlock()

    // Update Bitcoin node status
    if btcStatus, err := m.bitcoinNode.GetStatus(); err != nil {
        m.status["bitcoin"] = NodeStatus{
            NodeType:  "bitcoin",
            Connected: false,
            Error:     err.Error(),
        }
    } else {
        m.status["bitcoin"] = *btcStatus
    }

    // Update Ethereum node status
    if ethStatus, err := m.ethereumNode.GetStatus(); err != nil {
        m.status["ethereum"] = NodeStatus{
            NodeType:  "ethereum",
            Connected: false,
            Error:     err.Error(),
        }
    } else {
        m.status["ethereum"] = *ethStatus
    }
}

// ToJSON converts node status to JSON
func (s *NodeStatus) ToJSON() string {
    data, _ := json.Marshal(s)
    return string(data)
}
