# Certificate Management Tests

## C_CERTIFICATE Class Tests

  - **Environment Variable Setup**
    - [1.1.01] should have all required env variables for testnet
      ```pseudocode
      // 1. List all required environment variables
      requiredVars = [
          'TESTNET_CIRCULAR_MAIN_PUBLIC_CHAIN_ADDRESS',
          'TESTNET_CIRCULAR_MAIN_PUBLIC_ACCOUNT_PUBKEY',
          'TESTNET_CIRCULAR_MAIN_PUBLIC_ACCOUNT_PVTKEY',
          'TESTNET_CIRCULAR_MAIN_PUBLIC_ACCOUNT_SEED'
      ]
      
      // 2. Check each variable exists and is not a placeholder
      FOR each var IN requiredVars
          VERIFY process.env[var] IS_NOT_EMPTY
          VERIFY process.env[var] DOES_NOT_MATCH_PATTERN "placeholder"
      ```
  
  - **C_CERTIFICATE Class**
    - [1.1.02] should initialize with default values
      ```pseudocode
      // 1. Create new certificate instance
      certificate = NEW C_CERTIFICATE()
      
      // 2. Verify default values
      VERIFY certificate.data IS EMPTY
      VERIFY certificate.previousTxID IS EMPTY
      VERIFY certificate.previousBlock IS EMPTY
      VERIFY certificate.codeVersion IS LIB_VERSION
      ```
    - [1.1.03] shouldn't expose any methods publically
    
    - **set data method**
      - [1.1.04] should store data as hex (using librarys stringToHex)
        ```pseudocode
        // 1. Create certificate and test data
        certificate = NEW C_CERTIFICATE()
        testData = "test data is a string"
        
        // 2. Set data and verify hex conversion
        certificate.setData(testData)
        expectedHex = CONVERT_TO_HEX(testData)
        VERIFY certificate.data IS expectedHex
        ```
    
    - **get data method**
      - [1.1.05] should retrieve original data for simple strings
        ```pseudocode
        // 1. Setup
        certificate = NEW C_CERTIFICATE()
        originalData = "another test"
        
        // 2. Set and get data
        certificate.setData(originalData)
        VERIFY certificate.getData() IS originalData
        ```
      - [1.1.06] should return empty string if data is null or empty hex
        ```pseudocode
        // 1. Test null data
        certificate = NEW C_CERTIFICATE()
        EXPECT_ERROR WHEN certificate.getData(NULL) WITH_MESSAGE "Data is null"
        
        // 2. Test empty hex
        certificate.data = ""
        EXPECT_ERROR WHEN certificate.getData("") WITH_MESSAGE "Data is empty"
        ```
      - [1.1.07] should return empty string if data is "0x"
        ```pseudocode
        // 1. Setup
        certificate = NEW C_CERTIFICATE()
        certificate.data = "0x"
        
        // 2. Verify
        VERIFY certificate.getData() IS ""
        ```
      - [1.1.08] should correctly retrieve multi-byte unicode data
        ```pseudocode
        // 1. Setup
        certificate = NEW C_CERTIFICATE()
        unicodeData = "你好世界 😊"
        
        // 2. Test
        certificate.setData(unicodeData)
        VERIFY certificate.getData() IS unicodeData
        ```
    
    - **get JSON certificate method**
      - [1.1.09] should return a valid JSON string
        ```pseudocode
        // 1. Setup
        certificate = NEW C_CERTIFICATE()
        testData = "json test"
        certificate.setData(testData)
        certificate.previousTxID = "tx123"
        certificate.previousBlock = "block456"
        
        // 2. Get and verify JSON
        jsonCert = certificate.getJSONCertificate()
        parsedCert = JSON.parse(jsonCert)
        
        // 3. Verify structure
        VERIFY parsedCert HAS_PROPERTY "data"
        VERIFY parsedCert HAS_PROPERTY "previousTxID"
        VERIFY parsedCert HAS_PROPERTY "previousBlock"
        VERIFY parsedCert HAS_PROPERTY "version"
        VERIFY parsedCert.version IS LIB_VERSION
        ```
    
    - **get certificate size method**
      - [1.1.10] should return correct byte length
        ```pseudocode
        // 1. Setup
        certificate = NEW C_CERTIFICATE()
        testData = "size test"
        certificate.setData(testData)
        certificate.previousTxID = "txIDForSize"
        certificate.previousBlock = "blockIDForSize"
        
        // 2. Calculate expected size
        jsonString = JSON.stringify({
            data: CONVERT_TO_HEX(testData),
            previousTxID: "txIDForSize",
            previousBlock: "blockIDForSize",
            version: LIB_VERSION
        })
        expectedSize = BUFFER_BYTE_LENGTH(jsonString)
        
        // 3. Verify
        VERIFY certificate.getCertificateSize() IS expectedSize
        ``` 