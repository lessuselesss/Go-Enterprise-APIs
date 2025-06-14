# Utils Module Unit Tests

## Test Cases

### hex_fix
- [1.4.01] should remove '0x' prefix from hexadecimal string
  ```pseudocode
  // 1. Test setup
  test_data = "0x123abc"
  
  // 2. Test execution
  result = hex_fix(test_data)
  
  // 3. Test verification
  VERIFY result IS "123abc"
  ```
- [1.4.02] should handle strings without '0x' prefix
  ```pseudocode
  // 1. Test setup
  test_data = "123abc"
  
  // 2. Test execution
  result = hex_fix(test_data)
  
  // 3. Test verification
  VERIFY result IS "123abc"
  ```

### string_to_hex
- [1.4.03] should convert an ASCII string to its hexadecimal representation
  ```pseudocode
  // 1. Test setup
  test_data = "Hello, World!"
  expected = "48656c6c6f2c20576f726c6421"
  
  // 2. Test execution
  result = string_to_hex(test_data)
  
  // 3. Test verification
  VERIFY result IS expected
  ```

### hex_to_string
- [1.4.04] should convert a valid hexadecimal string to its ASCII representation
  ```pseudocode
  // 1. Test setup
  test_data = "48656c6c6f2c20576f726c6421"
  expected = "Hello, World!"
  
  // 2. Test execution
  result = hex_to_string(test_data)
  
  // 3. Test verification
  VERIFY result IS expected
  ```
- [1.4.05] should handle invalid hexadecimal input
  ```pseudocode
  // 1. Test setup
  test_data = "invalid_hex"
  
  // 2. Test execution
  EXPECT_ERROR WHEN hex_to_string(test_data)
  ```

### get_formatted_timestamp
- [1.4.06] should return a timestamp in the format YYYY:MM:DD-HH:mm:ss
  ```pseudocode
  // 1. Test execution
  timestamp = get_formatted_timestamp()
  
  // 2. Test verification
  VERIFY timestamp MATCHES_REGEX /^\d{4}:\d{2}:\d{2}-\d{2}:\d{2}:\d{2}$/
  
  // 3. Verify numeric values are within valid ranges
  parts = timestamp.split(/[:\-]/)
  year = CONVERT_TO_INT(parts[0])
  month = CONVERT_TO_INT(parts[1])
  day = CONVERT_TO_INT(parts[2])
  hour = CONVERT_TO_INT(parts[3])
  minute = CONVERT_TO_INT(parts[4])
  second = CONVERT_TO_INT(parts[5])
  
  VERIFY year IS_GREATER_THAN 2020
  VERIFY month IS_BETWEEN 1 AND 12
  VERIFY day IS_BETWEEN 1 AND 31
  VERIFY hour IS_BETWEEN 0 AND 23
  VERIFY minute IS_BETWEEN 0 AND 59
  VERIFY second IS_BETWEEN 0 AND 59
  ```

### certificate_size
- [1.4.07] should return the correct byte length for a given string
  ```pseudocode
  // 1. Test setup
  test_string_ascii = "hello"
  test_string_unicode = "你好"
  
  // 2. Test execution
  size_ascii = get_certificate_size(test_string_ascii)
  size_unicode = get_certificate_size(test_string_unicode)
  
  // 3. Test verification
  VERIFY size_ascii IS 5
  VERIFY size_unicode IS 6 // Each Chinese character is 3 bytes in UTF-8
  ``` 