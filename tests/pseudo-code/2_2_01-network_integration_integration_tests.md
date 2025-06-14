# Circular Enterprise APIs Test Suite

## 2. Integration Tests
### 2.1 Network Integration
- **CEP_Account Live Network Tests (against mainnet, testnet & devnet live networks)**
  - **open account method**
    - [2.1.1] should open account with valid address on real network
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      validAddress = "0x123..." // Real blockchain address
      
      // 2. Open account
      account.open(validAddress)
      
      // 3. Verify account state
      VERIFY account.address EQUALS validAddress
      VERIFY account.publicKey IS_NOT_EMPTY
      VERIFY account.info IS_NOT_EMPTY
      VERIFY account.blockchain EQUALS DEFAULT_CHAIN
      ```
    - [2.1.2] should handle invalid address formats on real network
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      invalidAddresses = [
          "0x",                    // Too short
          "0x123",                 // Invalid length
          "0xabcdefghijklmnop",   // Invalid characters
          "1234567890abcdef"      // Missing 0x prefix
      ]
      
      // 2. Test each invalid address
      FOR each addr IN invalidAddresses
          VERIFY CALLING account.open(addr) RAISES "Invalid address format"
      ```
    - [2.1.3] should maintain account state after opening on real network
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      validAddress = "0x123..."
      
      // 2. Open account
      account.open(validAddress)
      initialState = {
          address: account.address,
          publicKey: account.publicKey,
          info: account.info
      }
      
      // 3. Perform some operations
      account.setBlockchain("0xnewchain")
      account.setNetwork("testnet")
      
      // 4. Verify state maintained
      VERIFY account.address EQUALS initialState.address
      VERIFY account.publicKey EQUALS initialState.publicKey
      VERIFY account.info EQUALS initialState.info
      ```

  - **close account method**
    - [2.1.4] should close account and reset state on real network
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      account.open("0x123...")
      account.setBlockchain("0xchain")
      account.setNetwork("testnet")
      
      // 2. Close account
      account.close()
      
      // 3. Verify reset state
      VERIFY account.address IS EMPTY
      VERIFY account.publicKey IS EMPTY
      VERIFY account.info IS EMPTY
      VERIFY account.blockchain EQUALS DEFAULT_CHAIN
      VERIFY account.NAG_URL EQUALS DEFAULT_NAG
      VERIFY account.Nonce IS 0
      ```
    - [2.1.5] should handle closing non-existent account on real network
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      
      // 2. Close without opening
      account.close()
      
      // 3. Verify default state
      VERIFY account.address IS EMPTY
      VERIFY account.publicKey IS EMPTY
      VERIFY account.info IS EMPTY
      VERIFY account.blockchain EQUALS DEFAULT_CHAIN
      ```

  - **set blockchain method**
    - [2.1.6] should set blockchain and maintain state on real network
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      account.open("0x123...")
      newChain = "0xmynewchain"
      
      // 2. Set blockchain
      account.setBlockchain(newChain)
      
      // 3. Verify state
      VERIFY account.blockchain IS newChain
      VERIFY account.address IS NOT EMPTY
      VERIFY account.publicKey IS NOT EMPTY
      
      // 4. Verify network connectivity
      result = await account.updateAccount()
      VERIFY result IS TRUE
      ```
    - [2.1.7] should handle invalid blockchain IDs on real network
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      account.open("0x123...")
      invalidChains = [
          NULL,
          "",
          "invalid",
          "0x",           // Too short
          "0x123"         // Invalid format
      ]
      
      // 2. Test each invalid chain
      FOR each chain IN invalidChains
          VERIFY CALLING account.setBlockchain(chain) RAISES "Invalid blockchain ID"
      ```

  - **set network method**
    - [2.1.8] should set network and update NAG_URL for mainnet
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      account.open("0x123...")
      
      // 2. Set mainnet
      await account.setNetwork('mainnet')
      
      // 3. Verify configuration
      VERIFY account.NAG_URL IS "https://mainnet-nag.circularlabs.io/API/"
      VERIFY account.NETWORK_NODE IS_NOT_EMPTY
      
      // 4. Verify connectivity
      result = await account.updateAccount()
      VERIFY result IS TRUE
      ```
    - [2.1.9] should set network and update NAG_URL for testnet
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      account.open("0x123...")
      
      // 2. Set testnet
      await account.setNetwork('testnet')
      
      // 3. Verify configuration
      VERIFY account.NAG_URL IS "https://testnet-nag.circularlabs.io/API/"
      VERIFY account.NETWORK_NODE IS_NOT_EMPTY
      
      // 4. Verify connectivity
      result = await account.updateAccount()
      VERIFY result IS TRUE
      ```
    - [2.1.10] should set network and update NAG_URL for devnet
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      account.open("0x123...")
      
      // 2. Set devnet
      await account.setNetwork('devnet')
      
      // 3. Verify configuration
      VERIFY account.NAG_URL IS "https://devnet-nag.circularlabs.io/API/"
      VERIFY account.NETWORK_NODE IS_NOT_EMPTY
      
      // 4. Verify connectivity
      result = await account.updateAccount()
      VERIFY result IS TRUE
      ```
    - [2.1.11] should handle network connection failures
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      account.open("0x123...")
      
      // 2. Simulate network failure
      MOCK_NETWORK_FAILURE = TRUE
      
      // 3. Attempt network change
      VERIFY CALLING account.setNetwork('mainnet') RAISES "Network connection failed"
      
      // 4. Verify state unchanged
      VERIFY account.NAG_URL EQUALS DEFAULT_NAG
      ```

  - **update account method**
    - [2.1.12] should update account nonce on real network
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      account.open("0x123...")
      initialNonce = account.Nonce
      
      // 2. Update account
      result = await account.updateAccount()
      
      // 3. Verify nonce update
      VERIFY result IS TRUE
      VERIFY account.Nonce GREATER THAN initialNonce
      ```
    - [2.1.13] should maintain correct nonce sequence across multiple updates
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      account.open("0x123...")
      nonces = []
      
      // 2. Perform multiple updates
      FOR i = 1 TO 5
          await account.updateAccount()
          nonces.push(account.Nonce)
      
      // 3. Verify nonce sequence
      FOR i = 1 TO nonces.length - 1
          VERIFY nonces[i] EQUALS nonces[i-1] + 1
      ```
    - [2.1.14] should handle network errors during update
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      account.open("0x123...")
      initialNonce = account.Nonce
      
      // 2. Simulate network error
      MOCK_NETWORK_ERROR = TRUE
      
      // 3. Attempt update
      result = await account.updateAccount()
      
      // 4. Verify error handling
      VERIFY result IS FALSE
      VERIFY account.Nonce EQUALS initialNonce
      VERIFY account.lastError INCLUDES "Network error"
      ```

  - **sign data method**
    - [2.1.15] should sign data and verify signature on real network
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      account.open(realAddress)
      account.setNetwork("testnet")
      testData = "test data for signing"
      realPrivateKey = "0x1234567890abcdef"  // Real private key for testnet

      // 2. Test signature generation
      signature = account.signData(testData, realPrivateKey)

      // 3. Verify signature format
      VERIFY signature IS STRING
      VERIFY signature.length GREATER THAN 0
      VERIFY signature MATCHES /^[0-9a-f]+$/  // Should be hex

      // 4. Verify signature on real network
      verificationResult = await VERIFY_SIGNATURE_ON_NETWORK(
          testData,
          signature,
          account.publicKey,
          "testnet"
      )
      VERIFY verificationResult IS TRUE

      // 5. Test with different data types
      TEST_DATA_TYPES = [
          "simple string",
          "Hello 世界",  // Unicode
          "Special chars: !@#$%^&*()",
          "1234567890",
          "0x1234567890abcdef"
      ]

      FOR each data IN TEST_DATA_TYPES
          signature = account.signData(data, realPrivateKey)
          verificationResult = await VERIFY_SIGNATURE_ON_NETWORK(
              data,
              signature,
              account.publicKey,
              "testnet"
          )
          VERIFY verificationResult IS TRUE
      ```
    - [2.1.16] should handle signing with invalid private keys
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      account.open(realAddress)
      account.setNetwork("testnet")
      testData = "test data"

      // 2. Test invalid private key formats
      INVALID_PRIVATE_KEYS = [
          "0x",                    // Too short
          "0x123",                 // Invalid length
          "0xabcdefghijklmnop",   // Invalid characters
          "1234567890abcdef",     // Missing 0x prefix
          "0x" + "0".repeat(64),  // All zeros
          "0x" + "f".repeat(64),  // All ones
          null,                    // Null key
          undefined,              // Undefined key
          "",                     // Empty string
          "not_a_hex_string"      // Non-hex string
      ]

      // 3. Test each invalid key
      FOR each key IN INVALID_PRIVATE_KEYS
          try {
              account.signData(testData, key)
              VERIFY false  // Should not reach here
          } catch (error) {
              // Verify error
              VERIFY error.message INCLUDES "Invalid private key"
          }

      // 4. Test with malformed but valid-length keys
      MALFORMED_KEYS = [
          "0x" + "1".repeat(64),  // Valid length but invalid key
          "0x" + "a".repeat(64),  // Valid length but invalid key
          "0x" + "9".repeat(64)   // Valid length but invalid key
      ]

      FOR each key IN MALFORMED_KEYS
          try {
              account.signData(testData, key)
              VERIFY false  // Should not reach here
          } catch (error) {
              // Verify error
              VERIFY error.message INCLUDES "Invalid private key"
          }
      ```
    - [2.1.17] should maintain signature consistency across networks
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      account.open(realAddress)
      testData = "test data for consistency"
      realPrivateKey = "0x1234567890abcdef"
      networks = ["mainnet", "testnet", "devnet"]
      signatures = []

      // 2. Generate signatures on each network
      FOR each network IN networks
          account.setNetwork(network)
          signature = account.signData(testData, realPrivateKey)
          signatures.push(signature)

      // 3. Verify all signatures are identical
      FOR i = 1 TO signatures.length - 1
          VERIFY signatures[i] EQUALS signatures[0]

      // 4. Verify each signature on its respective network
      FOR i = 0 TO networks.length - 1
          verificationResult = await VERIFY_SIGNATURE_ON_NETWORK(
              testData,
              signatures[i],
              account.publicKey,
              networks[i]
          )
          VERIFY verificationResult IS TRUE

      // 5. Test with different data types across networks
      TEST_DATA_TYPES = [
          "simple string",
          "Hello 世界",  // Unicode
          "Special chars: !@#$%^&*()",
          "1234567890",
          "0x1234567890abcdef"
      ]

      FOR each data IN TEST_DATA_TYPES
          signatures = []
          FOR each network IN networks
              account.setNetwork(network)
              signature = account.signData(data, realPrivateKey)
              signatures.push(signature)
          
          // Verify consistency
          FOR i = 1 TO signatures.length - 1
              VERIFY signatures[i] EQUALS signatures[0]
          
          // Verify on each network
          FOR i = 0 TO networks.length - 1
              verificationResult = await VERIFY_SIGNATURE_ON_NETWORK(
                  data,
                  signatures[i],
                  account.publicKey,
                  networks[i]
              )
              VERIFY verificationResult IS TRUE

      // 6. Test signature uniqueness
      // Generate signatures for different data
      data1 = "test data 1"
      data2 = "test data 2"
      signature1 = account.signData(data1, realPrivateKey)
      signature2 = account.signData(data2, realPrivateKey)
      VERIFY signature1 NOT EQUALS signature2

      // 7. Test signature determinism
      // Generate multiple signatures for same data
      signature1 = account.signData(testData, realPrivateKey)
      signature2 = account.signData(testData, realPrivateKey)
      VERIFY signature1 EQUALS signature2
      ```

  - **get transaction method**
    - [2.1.18] should fetch a transaction by block number and ID on real network
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      account.open("0x123...")
      account.setNetwork("testnet")
      
      // 2. Submit a test transaction
      testData = "test transaction data"
      submitResult = await account.submitCertificate(testData, mockPrivateKey)
      
      // 3. Get transaction details
      txResult = await account.getTransaction(
          submitResult.blockNumber,
          submitResult.TxID
      )
      
      // 4. Verify transaction data
      VERIFY txResult.Result EQUALS 200
      VERIFY txResult.Response.id EQUALS submitResult.TxID
      VERIFY txResult.Response.status EQUALS "Confirmed"
      VERIFY txResult.Response.data EQUALS testData
      VERIFY txResult.Response.blockNumber EQUALS submitResult.blockNumber
      VERIFY txResult.Response.timestamp IS NOT NULL
      ```
    - [2.1.19] should handle non-existent transactions
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      account.open("0x123...")
      account.setNetwork("testnet")
      
      // 2. Attempt to get non-existent transaction
      nonExistentTxID = "0x" + "0".repeat(64)  // Valid format but non-existent
      blockNum = 1000000  // Far future block
      
      // 3. Get transaction details
      txResult = await account.getTransaction(blockNum, nonExistentTxID)
      
      // 4. Verify error handling
      VERIFY txResult.Result EQUALS 404
      VERIFY txResult.Message INCLUDES "Transaction Not Found"
      VERIFY txResult.Response IS NULL
      ```
    - [2.1.20] should handle invalid block numbers
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      account.open("0x123...")
      account.setNetwork("testnet")
      
      // 2. Test various invalid block numbers
      invalidBlocks = [
          -1,           // Negative block
          0,            // Zero block
          999999999,   // Far future block
          "invalid",    // Non-numeric
          NULL         // Null value
      ]
      
      // 3. Test each invalid block
      FOR each block IN invalidBlocks
          txResult = await account.getTransaction(block, "0x123...")
          VERIFY txResult.Result EQUALS 400
          VERIFY txResult.Message INCLUDES "Invalid block number"
          VERIFY txResult.Response IS NULL
      ```

  - **get transaction outcome method**
    - [2.1.21] should handle transaction polling and timeouts on real network
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      account.open(mockAddress)
      txID = "pollTxID456"
      account.intervalSec = 1
      
      // 2. Submit transaction and get outcome
      result = await account.submitCertificate("test data", mockPrivateKey)
      outcome = await account.getTransactionOutcome(result.TxID, 3)
      
      // 3. Verify outcome
      VERIFY outcome.Status EQUALS "Confirmed"
      VERIFY outcome.data EQUALS "test data"
      ```
    - [2.1.22] should handle pending transaction states
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      account.open(mockAddress)
      txID = "pendingTx789"
      account.intervalSec = 1
      
      // 2. Submit transaction
      result = await account.submitCertificate("pending test", mockPrivateKey)
      
      // 3. Get outcome with short timeout
      outcome = await account.getTransactionOutcome(result.TxID, 1)
      
      // 4. Verify pending state
      VERIFY outcome.Status EQUALS "Pending"
      ```
    - [2.1.23] should handle transaction not found scenarios
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      account.open(mockAddress)
      nonExistentTxID = "nonexistent123"
      
      // 2. Attempt to get outcome
      outcome = await account.getTransactionOutcome(nonExistentTxID, 2)
      
      // 3. Verify not found handling
      VERIFY outcome.Status EQUALS "Not Found"
      ```
    - [2.1.24] should validate transaction outcomes match submitted data
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      account.open(mockAddress)
      testData = "validation test data"
      
      // 2. Submit and verify
      result = await account.submitCertificate(testData, mockPrivateKey)
      outcome = await account.getTransactionOutcome(result.TxID, 5)
      
      // 3. Verify data integrity
      VERIFY outcome.Status EQUALS "Confirmed"
      VERIFY outcome.data EQUALS testData
      VERIFY outcome.timestamp IS NOT NULL
      VERIFY outcome.blockNumber IS NOT NULL
      ```

  - **submit certificate method**
    - [2.1.25] should submit a certificate successfully on real network
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      account.open("0x123...")
      account.setNetwork("testnet")
      testData = "test certificate data"
      initialNonce = account.Nonce
      
      // 2. Submit certificate
      result = await account.submitCertificate(testData, mockPrivateKey)
      
      // 3. Verify submission
      VERIFY result.Result EQUALS 200
      VERIFY result.TxID IS NOT NULL
      VERIFY result.Message EQUALS "Transaction Added"
      VERIFY account.LatestTxID EQUALS result.TxID
      VERIFY account.Nonce EQUALS initialNonce + 1
      ```
    - [2.1.26] should handle certificate submission with 1KB data
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      account.open("0x123...")
      account.setNetwork("testnet")
      
      // 2. Generate 1KB test data
      cert_1kb = generateCertificateWithSize(1024)
      
      // 3. Submit certificate
      result = await account.submitCertificate(cert_1kb, mockPrivateKey)
      
      // 4. Verify submission
      VERIFY result.Result EQUALS 200
      VERIFY result.TxID IS NOT NULL
      VERIFY result.Message EQUALS "Transaction Added"
      
      // 5. Verify data integrity
      txOutcome = await account.getTransactionOutcome(result.TxID, 5)
      VERIFY txOutcome.data EQUALS cert_1kb
      ```
    - [2.1.27] should handle certificate submission with 2KB data
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      account.open("0x123...")
      account.setNetwork("testnet")
      
      // 2. Generate 2KB test data
      cert_2kb = generateCertificateWithSize(2048)
      
      // 3. Submit certificate
      result = await account.submitCertificate(cert_2kb, mockPrivateKey)
      
      // 4. Verify submission
      VERIFY result.Result EQUALS 200
      VERIFY result.TxID IS NOT NULL
      VERIFY result.Message EQUALS "Transaction Added"
      
      // 5. Verify data integrity
      txOutcome = await account.getTransactionOutcome(result.TxID, 5)
      VERIFY txOutcome.data EQUALS cert_2kb
      ```
    - [2.1.28] should handle certificate submission with 5KB data
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      account.open("0x123...")
      account.setNetwork("testnet")
      
      // 2. Generate 5KB test data
      cert_5kb = generateCertificateWithSize(5120)
      
      // 3. Submit certificate
      result = await account.submitCertificate(cert_5kb, mockPrivateKey)
      
      // 4. Verify submission
      VERIFY result.Result EQUALS 200
      VERIFY result.TxID IS NOT NULL
      VERIFY result.Message EQUALS "Transaction Added"
      
      // 5. Verify data integrity
      txOutcome = await account.getTransactionOutcome(result.TxID, 5)
      VERIFY txOutcome.data EQUALS cert_5kb
      ```
    - [2.1.29] should handle concurrent certificate submissions
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      account.open("0x123...")
      account.setNetwork("testnet")
      initialNonce = account.Nonce
      
      // 2. Prepare multiple certificates
      certs = [
          "cert1",
          "cert2",
          "cert3"
      ]
      
      // 3. Submit concurrently
      results = await Promise.all(
          certs.map(cert => account.submitCertificate(cert, mockPrivateKey))
      )
      
      // 4. Verify all submissions
      FOR each result IN results
          VERIFY result.Result EQUALS 200
          VERIFY result.TxID IS NOT NULL
      
      // 5. Verify nonce sequence
      VERIFY account.Nonce EQUALS initialNonce + certs.length
      ```
    - [2.1.30] should handle network errors during submission
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      account.open("0x123...")
      account.setNetwork("testnet")
      initialNonce = account.Nonce
      
      // 2. Simulate network error
      MOCK_NETWORK_ERROR = TRUE
      
      // 3. Attempt submission
      result = await account.submitCertificate("test data", mockPrivateKey)
      
      // 4. Verify error handling
      VERIFY result.Result EQUALS 500
      VERIFY result.Message INCLUDES "Network error"
      VERIFY account.Nonce EQUALS initialNonce
      ```
    - [2.1.31] should maintain transaction order with multiple submissions
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      account.open("0x123...")
      account.setNetwork("testnet")
      initialNonce = account.Nonce
      
      // 2. Submit sequence of certificates
      certs = [
          "cert1",
          "cert2",
          "cert3"
      ]
      results = []
      
      FOR each cert IN certs
          result = await account.submitCertificate(cert, mockPrivateKey)
          results.push(result)
      
      // 3. Verify transaction order
      FOR i = 0 TO results.length - 1
          txOutcome = await account.getTransactionOutcome(results[i].TxID, 5)
          VERIFY txOutcome.data EQUALS certs[i]
          VERIFY txOutcome.nonce EQUALS initialNonce + i + 1
      ```

  - **network resilience tests**
    - [2.1.32] should handle network timeouts and retries
    - [2.1.33] should handle temporary network disconnections
    - [2.1.34] should handle rate limiting and backoff
    - [2.1.35] should maintain state during network issues 