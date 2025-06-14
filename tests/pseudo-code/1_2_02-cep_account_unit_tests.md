# Account Management Tests

## CEP_Account Class Tests

  - **CEP_Account Class**
    - [1.2.01] should initialize with default values
      ```pseudocode
      // 1. Create new account
      account = NEW CEP_Account()
      
      // 2. Verify default values
      VERIFY account.address IS EMPTY
      VERIFY account.publicKey IS EMPTY
      VERIFY account.info IS EMPTY
      VERIFY account.codeVersion IS LIB_VERSION
      VERIFY account.lastError IS ""
      VERIFY account.NAG_URL IS DEFAULT_NAG
      VERIFY account.NETWORK_NODE IS ""
      VERIFY account.blockchain IS DEFAULT_CHAIN
      VERIFY account.LatestTxID IS ""
      VERIFY account.Nonce IS 0
      VERIFY account.data IS EMPTY_OBJECT
      VERIFY account.intervalSec IS 2
      ```
    - [1.2.02] should only expose the open, set network, set blockchain, update account, submit certificate, get transaction outcome, get transaction & close methods publically.
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      
      // 2. Define allowed public methods (exactly as specified in requirements)
      ALLOWED_METHODS = [
          "open",
          "set network",
          "set blockchain",
          "update account",
          "submit certificate",
          "get transaction outcome",
          "get transaction",
          "close"
      ]
      
      // 3. Define known private methods (these should not be accessible)
      PRIVATE_METHODS = [
          "sign data",
          "verify signature",
          "validate address",
          "validate private key",
          "handle network error",
          "handle api error",
          "update nonce",
          "update latest tx id",
          "update account info"
      ]
      
      // 4. Define methods that should be removed/not accessible
      REMOVED_METHODS = [
          "get transaction by id"  // Explicitly test this method is not accessible
      ]
      
      // 5. Test public method accessibility
      FOR each method IN ALLOWED_METHODS
          VERIFY method IS_ACCESSIBLE_ON account
          VERIFY method IS_CALLABLE_ON account
      
      // 6. Test private method inaccessibility
      FOR each method IN PRIVATE_METHODS
          VERIFY method IS_NOT_ACCESSIBLE_ON account
          VERIFY method IS_NOT_CALLABLE_ON account
      
      // 7. Test removed method inaccessibility
      FOR each method IN REMOVED_METHODS
          VERIFY method IS_NOT_ACCESSIBLE_ON account
          VERIFY method IS_NOT_CALLABLE_ON account
          try {
              CALL account.get_transaction_by_id("test_id", 1, 100)
              VERIFY false  // Should not reach here
          } catch (error) {
              VERIFY error.message INCLUDES "not accessible"
          }
      
      // 8. Test method invocation
      // Test public method invocation
      try {
          CALL account.open("0x123...")
          VERIFY true  // Should not throw
      } catch (error) {
          VERIFY false  // Should not reach here
      }
      
      // Test private method invocation
      try {
          CALL account.sign_data("test", "0x123...")
          VERIFY false  // Should not reach here
      } catch (error) {
          VERIFY error.message INCLUDES "not accessible"
      }
      
      // 9. Test method enumeration
      methods = GET_ALL_METHODS(account)
      FOR each method IN methods
          IF method IN ALLOWED_METHODS
              VERIFY true  // Should be present
          ELSE IF method IN PRIVATE_METHODS
              VERIFY false  // Should not be present
          ELSE IF method IN REMOVED_METHODS
              VERIFY false  // Should not be present
          ELSE
              VERIFY false  // Unknown method found
      ```

    - **open account method**
      - [1.2.03] should set the account address
        ```pseudocode
        // 1. Setup
        account = NEW CEP_Account()
        mockAddress = "0x123..."
        
        // 2. Test
        account.open(mockAddress)
        VERIFY account.address IS mockAddress
        ```
      - [1.2.04] should throw an error for invalid address format
        ```pseudocode
        // 1. Setup
        account = NEW CEP_Account()
        
        // 2. Test invalid inputs
        EXPECT_ERROR WHEN account.open(NULL) WITH_MESSAGE "Invalid address format"
        EXPECT_ERROR WHEN account.open(123) WITH_MESSAGE "Invalid address format"
        EXPECT_ERROR WHEN account.open({}) WITH_MESSAGE "Invalid address format"
        ```

    - **close account method**
      - [1.2.05] should reset account properties to default
        ```pseudocode
        // 1. Setup
        account = NEW CEP_Account()
        account.open("0x123...")
        
        // 2. Test
        account.close()
        VERIFY account.address IS EMPTY
        VERIFY account.publicKey IS EMPTY
        VERIFY account.info IS EMPTY
        VERIFY account.lastError IS ""
        VERIFY account.NAG_URL IS DEFAULT_NAG
        VERIFY account.NETWORK_NODE IS ""
        VERIFY account.blockchain IS DEFAULT_CHAIN
        VERIFY account.LatestTxID IS ""
        VERIFY account.Nonce IS 0
        VERIFY account.data IS EMPTY_OBJECT
        VERIFY account.intervalSec IS 2
        ```

    - **set blockchain method**
      - [1.2.06] should update the blockchain property
        ```pseudocode
        // 1. Setup
        account = NEW CEP_Account()
        newChain = "0xmynewchain"
        
        // 2. Test
        account.setBlockchain(newChain)
        VERIFY account.blockchain IS newChain
        ```

    - **set network method**
      - [1.2.07] should update NAG_URL for "mainnet"
        ```pseudocode
        // 1. Setup
        account = NEW CEP_Account()
        account.open(mockAddress)
        initialNAG = account.NAG_URL
        
        // 2. Set mainnet
        await account.setNetwork('mainnet')
        
        // 3. Verify configuration
        VERIFY account.NAG_URL IS "https://mainnet-nag.circularlabs.io/API/"
        VERIFY account.NETWORK_NODE IS_NOT_EMPTY
        VERIFY account.NETWORK_NODE INCLUDES "mainnet"
        
        // 4. Verify account state
        VERIFY account.address IS mockAddress
        VERIFY account.publicKey IS_NOT_EMPTY
        VERIFY account.info IS_NOT_EMPTY
        VERIFY account.blockchain IS DEFAULT_CHAIN
        ```

      - [1.2.08] should update NAG_URL for "testnet"
        ```pseudocode
        // 1. Setup
        account = NEW CEP_Account()
        account.open(mockAddress)
        initialNAG = account.NAG_URL
        
        // 2. Set testnet
        await account.setNetwork('testnet')
        
        // 3. Verify configuration
        VERIFY account.NAG_URL IS "https://testnet-nag.circularlabs.io/API/"
        VERIFY account.NETWORK_NODE IS_NOT_EMPTY
        VERIFY account.NETWORK_NODE INCLUDES "testnet"
        
        // 4. Verify account state
        VERIFY account.address IS mockAddress
        VERIFY account.publicKey IS_NOT_EMPTY
        VERIFY account.info IS_NOT_EMPTY
        VERIFY account.blockchain IS DEFAULT_CHAIN
        ```

      - [1.2.09] should update NAG_URL for "devnet"
        ```pseudocode
        // 1. Setup
        account = NEW CEP_Account()
        account.open(mockAddress)
        initialNAG = account.NAG_URL
        
        // 2. Set devnet
        await account.setNetwork('devnet')
        
        // 3. Verify configuration
        VERIFY account.NAG_URL IS "https://devnet-nag.circularlabs.io/API/"
        VERIFY account.NETWORK_NODE IS_NOT_EMPTY
        VERIFY account.NETWORK_NODE INCLUDES "devnet"
        
        // 4. Verify account state
        VERIFY account.address IS mockAddress
        VERIFY account.publicKey IS_NOT_EMPTY
        VERIFY account.info IS_NOT_EMPTY
        VERIFY account.blockchain IS DEFAULT_CHAIN
        ```

      - [1.2.10] should throw an error if network request fails
        ```pseudocode
        // 1. Setup
        account = NEW CEP_Account()
        account.open(mockAddress)
        initialNAG = account.NAG_URL
        
        // 2. Test various network failure scenarios
        NETWORK_ERRORS = [
            { type: "timeout", message: "Request timeout" },
            { type: "connection", message: "Connection refused" },
            { type: "dns", message: "DNS lookup failed" },
            { type: "reset", message: "Connection reset" }
        ]
        
        // 3. Test each network error
        FOR each error IN NETWORK_ERRORS
            MOCK_NETWORK_ERROR = error
            
            // Attempt network change
            try {
                await account.setNetwork('mainnet')
                VERIFY false  // Should not reach here
            } catch (error) {
                // Verify error
                VERIFY error.message INCLUDES error.message
                VERIFY error.type IS "NetworkError"
                VERIFY account.NAG_URL IS initialNAG  // Should not change
            }
        
        // 4. Test partial network failure
        MOCK_PARTIAL_NETWORK_FAILURE = TRUE
        
        try {
            await account.setNetwork('mainnet')
            VERIFY false  // Should not reach here
        } catch (error) {
            VERIFY error.message INCLUDES "Connection lost"
            VERIFY error.type IS "NetworkError"
            VERIFY account.NAG_URL IS initialNAG  // Should not change
        }
        
        // 5. Verify account state remains unchanged
        VERIFY account.address IS mockAddress
        VERIFY account.publicKey IS_NOT_EMPTY
        VERIFY account.info IS_NOT_EMPTY
        VERIFY account.blockchain IS DEFAULT_CHAIN
        ```

      - [1.2.11] should throw an error if API response indicates failure
        ```pseudocode
        // 1. Setup
        account = NEW CEP_Account()
        account.open(mockAddress)
        initialNAG = account.NAG_URL
        
        // 2. Test various API error responses
        API_ERRORS = [
            { Result: 400, Message: "Bad Request" },
            { Result: 401, Message: "Unauthorized" },
            { Result: 403, Message: "Forbidden" },
            { Result: 404, Message: "Network Not Found" },
            { Result: 429, Message: "Too Many Requests" },
            { Result: 500, Message: "Internal Server Error" },
            { Result: 503, Message: "Service Unavailable" }
        ]
        
        // 3. Test each error case
        FOR each error IN API_ERRORS
            MOCK_API_RESPONSE = error
            
            // Attempt network change
            try {
                await account.setNetwork('mainnet')
                VERIFY false  // Should not reach here
            } catch (error) {
                // Verify error
                VERIFY error.message INCLUDES error.Message
                VERIFY error.type IS "APIError"
                VERIFY account.NAG_URL IS initialNAG  // Should not change
            }
        
        // 4. Test malformed API responses
        MALFORMED_RESPONSES = [
            { Result: 200 },  // Missing Response
            { Result: 200, Response: null },  // Null Response
            { Result: 200, Response: {} },  // Empty Response
            { Result: 200, Response: { Status: "Error" } }  // Error Status
        ]
        
        FOR each response IN MALFORMED_RESPONSES
            MOCK_API_RESPONSE = response
            
            try {
                await account.setNetwork('mainnet')
                VERIFY false  // Should not reach here
            } catch (error) {
                VERIFY error.message INCLUDES "Invalid response"
                VERIFY error.type IS "APIError"
                VERIFY account.NAG_URL IS initialNAG  // Should not change
            }
        
        // 5. Verify account state remains unchanged
        VERIFY account.address IS mockAddress
        VERIFY account.publicKey IS_NOT_EMPTY
        VERIFY account.info IS_NOT_EMPTY
        VERIFY account.blockchain IS DEFAULT_CHAIN
        ```

    - **update account method**
      - [1.2.12] should update Nonce on successful API call
        ```pseudocode
        // 1. Setup
        account = NEW CEP_Account()
        account.open(mockAddress)
        account.setNetwork("testnet")
        initialNonce = account.Nonce

        // 2. Mock successful API response
        MOCK_API_RESPONSE = { 
            Result: 200, 
            Response: { 
                Nonce: 5,
                Address: account.address,
                PublicKey: account.publicKey,
                LatestTxID: "0x123...",
                Status: "Active"
            } 
        }

        // 3. Update account
        result = await account.updateAccount()

        // 4. Verify successful update
        VERIFY result IS TRUE
        VERIFY account.Nonce EQUALS 6  // Should increment by 1
        VERIFY account.LatestTxID EQUALS MOCK_API_RESPONSE.Response.LatestTxID
        VERIFY account.info.Status EQUALS "Active"

        // 5. Verify account state
        VERIFY account.address EQUALS MOCK_API_RESPONSE.Response.Address
        VERIFY account.publicKey EQUALS MOCK_API_RESPONSE.Response.PublicKey
        ```

    - **sign data method**
      - [1.2.17] should throw an error if account is not open
        ```pseudocode
        // 1. Setup
        account = NEW CEP_Account()
        
        // 2. Test
        EXPECT_ERROR WHEN account.sign_data("test data", "0x11223344556677889900aabbccddeeff11223344556677889900aabbccddeeff") WITH_MESSAGE "Account not open"
        ```

      - [1.2.18] should produce different signatures for different data
        ```pseudocode
        // 1. Setup
        account = NEW CEP_Account()
        account.open("0x123...")
        private_key = "0x11223344556677889900aabbccddeeff11223344556677889900aabbccddeeff"
        
        // 2. Test
        signature1 = account.sign_data("data1", private_key)
        signature2 = account.sign_data("data2", private_key)
        
        // 3. Verify
        VERIFY signature1 IS_NOT signature2
        ```

      - [1.2.19] should produce different signatures for different private keys
        ```pseudocode
        // 1. Setup
        account = NEW CEP_Account()
        account.open("0x123...")
        private_key1 = "0x11223344556677889900aabbccddeeff11223344556677889900aabbccddeeff"
        private_key2 = "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
        
        // 2. Test
        signature1 = account.sign_data("data", private_key1)
        signature2 = account.sign_data("data", private_key2)
        
        // 3. Verify
        VERIFY signature1 IS_NOT signature2
        ```

    - **submit certificate method**
      - [1.2.25] should successfully submit a certificate
        ```pseudocode
        // 1. Setup
        account = NEW CEP_Account()
        account.open("0x123...")
        account.setNetwork("testnet")
        
        // 2. Mock successful API response
        MOCK_API_RESPONSE = { 
            Result: 200, 
            Response: { 
                txId: "0xabcdef1234567890"
            } 
        }
        
        // 3. Submit certificate
        result = await account.submit_certificate("test data", "0x11223344556677889900aabbccddeeff11223344556677889900aabbccddeeff")
        
        // 4. Verify successful submission
        VERIFY result IS TRUE
        VERIFY account.nonce EQUALS 1
        VERIFY account.latest_tx_id EQUALS "0xabcdef1234567890"
        ```

    - **network resilience**
      - [1.2.35] should preserve state during network issues
        ```pseudocode
        // 1. Setup
        account = NEW CEP_Account()
        account.open("0x123...")
        account.setNetwork("testnet")
        initial_nonce = account.nonce
        
        // 2. Mock network error
        MOCK_NETWORK_ERROR = { type: "timeout", message: "Request timeout" }
        
        // 3. Attempt to update account
        result = await account.update_account()
        
        // 4. Verify state preservation
        VERIFY result IS FALSE
        VERIFY account.nonce EQUALS initial_nonce
        ```

      - [1.2.13] should return false and not update Nonce on API error (Result != 200)
        ```pseudocode
        // 1. Setup
        account = NEW CEP_Account()
        account.open(mockAddress)
        account.setNetwork("testnet")
        initialNonce = account.Nonce

        // 2. Test various API error responses
        API_ERRORS = [
            { Result: 400, Message: "Bad Request" },
            { Result: 401, Message: "Unauthorized" },
            { Result: 403, Message: "Forbidden" },
            { Result: 404, Message: "Account Not Found" },
            { Result: 429, Message: "Too Many Requests" },
            { Result: 500, Message: "Internal Server Error" },
            { Result: 503, Message: "Service Unavailable" }
        ]

        // 3. Test each error case
        FOR each error IN API_ERRORS
            MOCK_API_RESPONSE = error
            
            // Attempt update
            result = await account.updateAccount()
            
            // Verify error handling
            VERIFY result IS FALSE
            VERIFY account.Nonce EQUALS initialNonce  // Nonce should not change
            VERIFY account.lastError INCLUDES error.Message
            VERIFY account.lastError INCLUDES error.Result.toString()

        // 4. Verify account state remains unchanged
        VERIFY account.address EQUALS mockAddress
        VERIFY account.publicKey IS_NOT_EMPTY
        VERIFY account.info IS_NOT_EMPTY
        ```

      - [1.2.14] should return false on network error
        ```pseudocode
        // 1. Setup
        account = NEW CEP_Account()
        account.open(mockAddress)
        account.setNetwork("testnet")
        initialNonce = account.Nonce

        // 2. Test various network error scenarios
        NETWORK_ERRORS = [
            { type: "timeout", message: "Request timeout" },
            { type: "connection", message: "Connection refused" },
            { type: "dns", message: "DNS lookup failed" },
            { type: "reset", message: "Connection reset" }
        ]

        // 3. Test each network error
        FOR each error IN NETWORK_ERRORS
            MOCK_NETWORK_ERROR = error
            
            // Attempt update
            result = await account.updateAccount()
            
            // Verify error handling
            VERIFY result IS FALSE
            VERIFY account.Nonce EQUALS initialNonce  // Nonce should not change
            VERIFY account.lastError INCLUDES error.message
            VERIFY account.lastError INCLUDES "Network error"

        // 4. Test partial network failure
        MOCK_PARTIAL_NETWORK_FAILURE = TRUE
        result = await account.updateAccount()
        VERIFY result IS FALSE
        VERIFY account.Nonce EQUALS initialNonce
        VERIFY account.lastError INCLUDES "Connection lost"
        ```

      - [1.2.15] should throw an error if account is not open
        ```pseudocode
        // 1. Setup
        account = NEW CEP_Account()

        // 2. Attempt update without opening account
        EXPECT_ERROR WHEN account.updateAccount() WITH {
            message: "Account not open",
            type: "AccountError"
        }

        // 3. Verify account remains in default state
        VERIFY account.address IS EMPTY
        VERIFY account.publicKey IS EMPTY
        VERIFY account.info IS EMPTY
        VERIFY account.Nonce IS 0
        VERIFY account.LatestTxID IS ""
        VERIFY account.lastError IS ""
        ```

      - [1.2.16] should return false if response is malformed (missing Nonce)
        ```pseudocode
        // 1. Setup
        account = NEW CEP_Account()
        account.open(mockAddress)
        account.setNetwork("testnet")
        initialNonce = account.Nonce

        // 2. Test various malformed responses
        MALFORMED_RESPONSES = [
            { Result: 200, Response: {} },  // Empty response
            { Result: 200, Response: { Address: "0x123..." } },  // Missing Nonce
            { Result: 200, Response: { Nonce: "invalid" } },  // Invalid Nonce type
            { Result: 200, Response: { Nonce: -1 } },  // Negative Nonce
            { Result: 200, Response: { Nonce: null } },  // Null Nonce
            { Result: 200, Response: { Nonce: undefined } },  // Undefined Nonce
            { Result: 200 },  // Missing Response
            { Result: 200, Response: null },  // Null Response
            { Result: 200, Response: undefined }  // Undefined Response
        ]

        // 3. Test each malformed response
        FOR each response IN MALFORMED_RESPONSES
            MOCK_API_RESPONSE = response
            
            // Attempt update
            result = await account.updateAccount()
            
            // Verify error handling
            VERIFY result IS FALSE
            VERIFY account.Nonce EQUALS initialNonce  // Nonce should not change
            VERIFY account.lastError INCLUDES "Invalid response"
            VERIFY account.lastError INCLUDES "Nonce"

        // 4. Verify account state remains unchanged
        VERIFY account.address EQUALS mockAddress
        VERIFY account.publicKey IS_NOT_EMPTY
        VERIFY account.info IS_NOT_EMPTY
        ```

    - **sign data method**
      - [1.2.17] should throw an error if account is not open
        ```pseudocode
        // 1. Setup
        account = NEW CEP_Account()
        testData = "test data for signing"
        testPrivateKey = "0x1234567890abcdef"

        // 2. Attempt to sign without opening account
        EXPECT_ERROR WHEN account.signData(testData, testPrivateKey) WITH {
            message: "Account not open",
            type: "AccountError"
        }

        // 3. Verify account remains in default state
        VERIFY account.address IS EMPTY
        VERIFY account.publicKey IS EMPTY
        VERIFY account.info IS EMPTY
        VERIFY account.Nonce IS 0
        VERIFY account.LatestTxID IS ""
        ```

      - [1.2.18] should produce different signatures for different data
        ```pseudocode
        // 1. Setup
        account = NEW CEP_Account()
        account.open(mockAddress)
        account.setNetwork("testnet")
        testPrivateKey = "0x1234567890abcdef"

        // 2. Test with different data types
        TEST_DATA = [
            "simple string",
            "Hello 世界",  // Unicode
            "Special chars: !@#$%^&*()",
            "1234567890",
            "0x1234567890abcdef",
            "",  // Empty string
            "   ",  // Whitespace
            "a".repeat(1000)  // Long string
        ]

        // 3. Generate signatures for each data type
        signatures = []
        FOR each data IN TEST_DATA
            signature = account.signData(data, testPrivateKey)
            signatures.push(signature)
            
            // Verify signature format
            VERIFY signature IS STRING
            VERIFY signature.length GREATER THAN 0
            VERIFY signature MATCHES /^[0-9a-f]+$/  // Should be hex

        // 4. Verify all signatures are unique
        FOR i = 0 TO signatures.length - 1
            FOR j = i + 1 TO signatures.length - 1
                VERIFY signatures[i] NOT EQUALS signatures[j]

        // 5. Verify signatures are deterministic
        FOR each data IN TEST_DATA
            signature1 = account.signData(data, testPrivateKey)
            signature2 = account.signData(data, testPrivateKey)
            VERIFY signature1 EQUALS signature2
        ```

      - [1.2.19] should produce different signatures for different private keys
        ```pseudocode
        // 1. Setup
        account = NEW CEP_Account()
        account.open(mockAddress)
        account.setNetwork("testnet")
        testData = "test data for signing"

        // 2. Test with different private keys
        TEST_PRIVATE_KEYS = [
            "0x" + "1".repeat(64),  // All ones
            "0x" + "a".repeat(64),  // All a's
            "0x" + "f".repeat(64),  // All f's
            "0x" + "0".repeat(64),  // All zeros
            "0x" + "5".repeat(64),  // All fives
            "0x" + "9".repeat(64)   // All nines
        ]

        // 3. Generate signatures for each private key
        signatures = []
        FOR each key IN TEST_PRIVATE_KEYS
            signature = account.signData(testData, key)
            signatures.push(signature)
            
            // Verify signature format
            VERIFY signature IS STRING
            VERIFY signature.length GREATER THAN 0
            VERIFY signature MATCHES /^[0-9a-f]+$/  // Should be hex

        // 4. Verify all signatures are unique
        FOR i = 0 TO signatures.length - 1
            FOR j = i + 1 TO signatures.length - 1
                VERIFY signatures[i] NOT EQUALS signatures[j]

        // 5. Test with invalid private keys
        INVALID_KEYS = [
            "0x",                    // Too short
            "0x123",                 // Invalid length
            "0xabcdefghijklmnop",   // Invalid characters
            "1234567890abcdef",     // Missing 0x prefix
            null,                    // Null key
            undefined,              // Undefined key
            "",                     // Empty string
            "not_a_hex_string"      // Non-hex string
        ]

        FOR each key IN INVALID_KEYS
            EXPECT_ERROR WHEN account.signData(testData, key) WITH {
                message: "Invalid private key",
                type: "ValidationError"
            }

        // 6. Verify signature verification
        FOR each key IN TEST_PRIVATE_KEYS
            signature = account.signData(testData, key)
            verificationResult = VERIFY_SIGNATURE(testData, signature, account.publicKey)
            VERIFY verificationResult IS TRUE
        ```

    - **get transaction and get transaction by ID methods**
      - [1.2.20] get transaction(BlockID, TxID) should fetch a transaction
        ```pseudocode
        // 1. Setup
        account = NEW CEP_Account()
        account.open(mockAddress)
        account.setNetwork("testnet")
        
        // 2. Submit a test transaction first
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
        VERIFY txResult.Response.status IS_NOT "Pending"
        VERIFY txResult.Response.data EQUALS testData
        VERIFY txResult.Response.blockNumber EQUALS submitResult.blockNumber
        VERIFY txResult.Response.timestamp IS NOT NULL
        
        // 5. Verify transaction metadata
        VERIFY txResult.Response.signature IS NOT NULL
        VERIFY txResult.Response.publicKey EQUALS account.publicKey
        VERIFY txResult.Response.nonce EQUALS account.Nonce - 1  // Previous nonce
        ```

      - [1.2.21] get transaction(BlockID, TxID) should throw on network error
        ```pseudocode
        // 1. Setup
        account = NEW CEP_Account()
        account.open(mockAddress)
        account.setNetwork("testnet")
        blockNum = 100
        txID = "testTxID123"
        
        // 2. Simulate network errors
        NETWORK_ERRORS = [
            { type: "timeout", message: "Request timeout" },
            { type: "connection", message: "Connection refused" },
            { type: "dns", message: "DNS lookup failed" }
        ]
        
        // 3. Test each network error
        FOR each error IN NETWORK_ERRORS
            MOCK_NETWORK_ERROR = error
            
            // Attempt to get transaction
            try {
                await account.getTransaction(blockNum, txID)
                VERIFY false  // Should not reach here
            } catch (error) {
                // Verify error
                VERIFY error.message INCLUDES error.message
                VERIFY error.type IS "NetworkError"
            }
        
        // 4. Verify account state remains unchanged
        VERIFY account.Nonce EQUALS initialNonce
        VERIFY account.LatestTxID IS ""
        ```

      - [1.2.22] get transaction by ID should fetch a transaction within a block range
        ```pseudocode
        // 1. Setup
        account = NEW CEP_Account()
        account.open(mockAddress)
        account.setNetwork("testnet")
        
        // 2. Submit multiple test transactions
        testData = ["tx1", "tx2", "tx3"]
        submitResults = []
        
        FOR each data IN testData
            result = await account.submitCertificate(data, mockPrivateKey)
            submitResults.push(result)
        
        // 3. Get transactions within block range
        startBlock = submitResults[0].blockNumber
        endBlock = submitResults[2].blockNumber
        
        FOR each result IN submitResults
            txResult = await account.getTransactionByID(
                result.TxID,
                startBlock,
                endBlock
            )
            
            // Verify transaction data
            VERIFY txResult.Result EQUALS 200
            VERIFY txResult.Response.id EQUALS result.TxID
            VERIFY txResult.Response.status IS_NOT "Pending"
            VERIFY txResult.Response.blockNumber GREATER THAN OR EQUAL TO startBlock
            VERIFY txResult.Response.blockNumber LESS THAN OR EQUAL TO endBlock
            VERIFY txResult.Response.data EQUALS testData[submitResults.indexOf(result)]
        
        // 4. Test block range edge cases
        // Test with exact block numbers
        txResult = await account.getTransactionByID(
            submitResults[0].TxID,
            submitResults[0].blockNumber,
            submitResults[0].blockNumber
        )
        VERIFY txResult.Result EQUALS 200
        VERIFY txResult.Response.status IS_NOT "Pending"
        
        // Test with invalid block range
        try {
            await account.getTransactionByID(
                submitResults[0].TxID,
                endBlock,
                startBlock  // Reversed range
            )
            VERIFY false  // Should not reach here
        } catch (error) {
            VERIFY error.message INCLUDES "Invalid block range"
        }
        ```

      - [1.2.23] get transaction by ID should handle "Transaction Not Found"
        ```pseudocode
        // 1. Setup
        account = NEW CEP_Account()
        account.open(mockAddress)
        account.setNetwork("testnet")
        
        // 2. Test various "not found" scenarios
        NOT_FOUND_SCENARIOS = [
            { txID: "0x" + "0".repeat(64), blockRange: [1, 100] },  // Non-existent TxID
            { txID: "0x123", blockRange: [1, 100] },                // Invalid TxID format
            { txID: "0x" + "f".repeat(64), blockRange: [999999, 1000000] }  // Far future block
        ]
        
        // 3. Test each scenario
        FOR each scenario IN NOT_FOUND_SCENARIOS
            txResult = await account.getTransactionByID(
                scenario.txID,
                scenario.blockRange[0],
                scenario.blockRange[1]
            )
            
            // Verify not found response
            VERIFY txResult.Result EQUALS 404
            VERIFY txResult.Message INCLUDES "Transaction Not Found"
            VERIFY txResult.Response IS NULL
        
        // 4. Test with valid TxID but wrong block range
        // First submit a transaction
        submitResult = await account.submitCertificate("test data", mockPrivateKey)
        
        // Try to find it in wrong block range
        txResult = await account.getTransactionByID(
            submitResult.TxID,
            submitResult.blockNumber + 1000,  // Far future block
            submitResult.blockNumber + 2000
        )
        
        VERIFY txResult.Result EQUALS 404
        VERIFY txResult.Message INCLUDES "Transaction Not Found"
        ```

      - [1.2.24] get transaction by ID should throw on network error
        ```pseudocode
        // 1. Setup
        account = NEW CEP_Account()
        account.open(mockAddress)
        account.setNetwork("testnet")
        txID = "testTxID123"
        
        // 2. Simulate various network error scenarios
        NETWORK_ERRORS = [
            { type: "timeout", message: "Request timeout", retryCount: 3 },
            { type: "connection", message: "Connection refused", retryCount: 2 },
            { type: "dns", message: "DNS lookup failed", retryCount: 1 }
        ]
        
        // 3. Test each network error with retries
        FOR each error IN NETWORK_ERRORS
            MOCK_NETWORK_ERROR = error
            retryAttempts = 0
            
            // Attempt to get transaction with retries
            try {
                await account.getTransactionByID(txID, 1, 100)
                VERIFY false  // Should not reach here
            } catch (error) {
                // Verify error and retry behavior
                VERIFY error.message INCLUDES error.message
                VERIFY error.type IS "NetworkError"
                VERIFY retryAttempts EQUALS error.retryCount
            }
        
        // 4. Test partial network failure
        // Simulate network failure after successful connection
        MOCK_PARTIAL_NETWORK_FAILURE = TRUE
        
        try {
            await account.getTransactionByID(txID, 1, 100)
            VERIFY false  // Should not reach here
        } catch (error) {
            VERIFY error.message INCLUDES "Connection lost"
            VERIFY error.type IS "NetworkError"
        }
        
        // 5. Verify account state remains unchanged after errors
        VERIFY account.Nonce EQUALS initialNonce
        VERIFY account.LatestTxID IS ""
        ```

    - **submit certificate method**
      - [1.2.25] should submit a certificate successfully on real network
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
      - [1.2.26] should handle certificate submission with 1KB data
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
      - [1.2.27] should handle certificate submission with 2KB data
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
      - [1.2.28] should handle certificate submission with 5KB data
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
      - [1.2.29] should handle concurrent certificate submissions
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
      - [1.2.30] should handle network errors during submission
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
      - [1.2.31] should maintain transaction order with multiple submissions
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
    - [1.2.32] should handle network timeouts and retries
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      account.open(mockAddress)
      account.setNetwork("testnet")
      initialNonce = account.Nonce

      // 2. Test various timeout scenarios
      TIMEOUT_SCENARIOS = [
          { timeout: 1000, retries: 3 },  // Short timeout, multiple retries
          { timeout: 5000, retries: 2 },  // Medium timeout, fewer retries
          { timeout: 10000, retries: 1 }  // Long timeout, single retry
      ]

      // 3. Test each timeout scenario
      FOR each scenario IN TIMEOUT_SCENARIOS
          MOCK_TIMEOUT = scenario.timeout
          MOCK_RETRIES = scenario.retries
          retryAttempts = 0

          // Attempt operation
          try {
              await account.updateAccount()
              VERIFY false  // Should not reach here
          } catch (error) {
              // Verify retry behavior
              VERIFY retryAttempts EQUALS scenario.retries
              VERIFY error.message INCLUDES "Timeout"
              VERIFY error.type IS "NetworkError"
          }

      // 4. Test exponential backoff
      MOCK_TIMEOUT = 1000
      MOCK_RETRIES = 3
      retryDelays = []

      try {
          await account.updateAccount()
          VERIFY false  // Should not reach here
      } catch (error) {
          // Verify backoff timing
          FOR i = 1 TO retryDelays.length - 1
              VERIFY retryDelays[i] GREATER THAN retryDelays[i-1]
      }
      ```

    - [1.2.33] should handle temporary network disconnections
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      account.open(mockAddress)
      account.setNetwork("testnet")
      initialNonce = account.Nonce

      // 2. Test various disconnection scenarios
      DISCONNECTION_SCENARIOS = [
          { duration: 1000, shouldRecover: true },   // Short disconnection
          { duration: 5000, shouldRecover: true },   // Medium disconnection
          { duration: 10000, shouldRecover: false }  // Long disconnection
      ]

      // 3. Test each disconnection scenario
      FOR each scenario IN DISCONNECTION_SCENARIOS
          MOCK_DISCONNECTION = scenario.duration
          
          // Attempt operation
          try {
              await account.updateAccount()
              VERIFY scenario.shouldRecover  // Should only succeed if shouldRecover is true
          } catch (error) {
              VERIFY NOT scenario.shouldRecover  // Should only fail if shouldRecover is false
              VERIFY error.message INCLUDES "Connection lost"
              VERIFY error.type IS "NetworkError"
          }

      // 4. Test reconnection behavior
      MOCK_DISCONNECTION = 2000
      MOCK_RECONNECTION_DELAY = 1000

      // First attempt should fail
      try {
          await account.updateAccount()
          VERIFY false  // Should not reach here
      } catch (error) {
          VERIFY error.message INCLUDES "Connection lost"
      }

      // Wait for recovery
      await WAIT(MOCK_RECONNECTION_DELAY)

      // Second attempt should succeed
      result = await account.updateAccount()
      VERIFY result IS TRUE
      ```

    - [1.2.34] should handle rate limiting and backoff
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      account.open(mockAddress)
      account.setNetwork("testnet")
      initialNonce = account.Nonce

      // 2. Test rate limiting scenarios
      RATE_LIMIT_SCENARIOS = [
          { limit: 10, window: 1000 },  // 10 requests per second
          { limit: 100, window: 60000 } // 100 requests per minute
      ]

      // 3. Test each rate limit scenario
      FOR each scenario IN RATE_LIMIT_SCENARIOS
          MOCK_RATE_LIMIT = scenario
          requestCount = 0
          startTime = Date.now()

          // Make multiple requests
          FOR i = 1 TO scenario.limit + 5  // Exceed the limit
              try {
                  await account.updateAccount()
                  requestCount++
              } catch (error) {
                  // Verify rate limit error
                  VERIFY error.message INCLUDES "Rate limit exceeded"
                  VERIFY error.type IS "RateLimitError"
                  break
              }

          // Verify rate limit behavior
          elapsedTime = Date.now() - startTime
          VERIFY requestCount LESS THAN OR EQUAL TO scenario.limit
          VERIFY elapsedTime LESS THAN scenario.window

      // 4. Test backoff behavior
      MOCK_RATE_LIMIT = { limit: 5, window: 1000 }
      backoffDelays = []

      // Exceed rate limit
      FOR i = 1 TO 10
          try {
              await account.updateAccount()
          } catch (error) {
              backoffDelays.push(error.retryAfter)
          }

      // Verify increasing backoff
      FOR i = 1 TO backoffDelays.length - 1
          VERIFY backoffDelays[i] GREATER THAN backoffDelays[i-1]
      ```

    - [1.2.35] should maintain state during network issues
      ```pseudocode
      // 1. Setup
      account = NEW CEP_Account()
      account.open(mockAddress)
      account.setNetwork("testnet")
      initialNonce = account.Nonce
      initialState = {
          address: account.address,
          publicKey: account.publicKey,
          info: account.info,
          blockchain: account.blockchain
      }

      // 2. Test state preservation during various network issues
      NETWORK_ISSUES = [
          { type: "timeout", duration: 5000 },
          { type: "disconnection", duration: 3000 },
          { type: "rate_limit", duration: 2000 },
          { type: "partial_failure", duration: 1000 }
      ]

      // 3. Test each network issue
      FOR each issue IN NETWORK_ISSUES
          MOCK_NETWORK_ISSUE = issue
          
          // Attempt operation
          try {
              await account.updateAccount()
          } catch (error) {
              // Verify state preservation
              VERIFY account.address EQUALS initialState.address
              VERIFY account.publicKey EQUALS initialState.publicKey
              VERIFY account.info EQUALS initialState.info
              VERIFY account.blockchain EQUALS initialState.blockchain
          }

      // 4. Test state recovery after network issues
      MOCK_NETWORK_ISSUE = { type: "disconnection", duration: 2000 }
      
      // First attempt should fail
      try {
          await account.updateAccount()
          VERIFY false  // Should not reach here
      } catch (error) {
          // Verify state preserved
          VERIFY account.address EQUALS initialState.address
      }

      // Wait for recovery
      await WAIT(2000)

      // Second attempt should succeed
      result = await account.updateAccount()
      VERIFY result IS TRUE
      VERIFY account.Nonce GREATER THAN initialNonce
      ``` 