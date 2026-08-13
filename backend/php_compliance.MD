## 1. Core Server & HTTP Routing

In Go, we use the standard `net/http` package to listen for requests, read cookies, and send JSON. In PHP, this is handled automatically by the native CGI/SAPI environment and superglobals.

| Go Implementation (`net/http`) | PHP Standard Library Equivalent | Explanation |
| --- | --- | --- |
| `http.Server` / `http.ServeMux` | `php -S` / `$_SERVER['REQUEST_URI']` | PHP natively maps URLs to scripts or uses the `$_SERVER` array for basic manual routing. |
| `w.Header().Set("Content-Type", ...)` | `header('Content-Type: application/json')` | Native HTTP header manipulation. |
| `http.SetCookie(...)` | `setcookie(...)` | Native cookie handling. |
| `r.Cookie("access_token")` | `$_COOKIE['access_token']` | Accessing parsed incoming cookies. |
| `http.Redirect(...)` | `header("Location: /docs", true, 301)` | Standard HTTP redirect headers. |

*(Note: `oapi-codegen` was used purely as an offline build tool to generate standard `net/http` boilerplate. It is not a runtime framework.)*

## 2. Database & Queries

We use Go's standard `database/sql` package to interact with the database. The `lib/pq` package is solely a driver that speaks the PostgreSQL wire protocol, just like PHP's native PDO driver. No ORMs are used.

| Go Implementation (`database/sql`) | PHP Standard Library Equivalent | Explanation |
| --- | --- | --- |
| `sql.Open("postgres", dbUrl)` | `new PDO("pgsql:host=...")` | Standard relational database connection. |
| `db.QueryRowContext(...)` | `$stmt = $pdo->prepare(); $stmt->execute();` | Prepared statements to prevent SQL injection. |
| `row.Scan(&id, &username)` | `$row = $stmt->fetch(PDO::FETCH_ASSOC);` | Mapping raw SQL row returns to variables. |

*(Note: `sqlc` is used as an offline code generator to write standard `database/sql` boilerplate. It does not exist at runtime and is not an ORM).*

## 3. Security & Password Hashing

Storing plaintext passwords is a critical failure. Go provides a `bcrypt` implementation in its official `x/crypto` extension repository, exactly mirroring PHP's built-in password hashing suite.

| Go Implementation (`x/crypto/bcrypt`) | PHP Standard Library Equivalent | Explanation |
| --- | --- | --- |
| `bcrypt.GenerateFromPassword(...)` | `password_hash($pass, PASSWORD_BCRYPT)` | Secure, salted password hashing using the industry-standard Blowfish cipher. |
| `bcrypt.CompareHashAndPassword(...)` | `password_verify($pass, $hash)` | Securely comparing a plaintext string against a bcrypt hash. |

## 4. JWT & Token Generation

JSON Web Tokens (JWT) and Refresh Tokens rely on cryptographic randomness and HMAC hashing. While Go uses `golang-jwt/jwt/v5` for convenience to assemble the strings, the underlying mathematics are completely native to PHP.

| Go Implementation (`crypto` / `jwt`) | PHP Standard Library Equivalent | Explanation |
| --- | --- | --- |
| `jwt.NewWithClaims(SigningMethodHS256)` | `hash_hmac('sha256', $header.$payload, $secret)` | A JWT is simply a base64 encoded string signed with an HMAC SHA-256 hash. PHP has this natively built-in via `hash_hmac`. |
| `rand.Read(...)` | `random_bytes(32)` | Cryptographically secure pseudo-random number generation (CSPRNG) for refresh tokens. |
| `base64.URLEncoding.EncodeToString()` | `base64_encode()` (with string replacement) | Encoding binary tokens into safe URL strings. |
| `sha256.Sum256(...)` | `hash('sha256', $data)` | Hashing refresh tokens before storing them at rest. |

## 5. Data Serialization

Parsing incoming request bodies and formatting outgoing API responses.

| Go Implementation (`encoding/json`) | PHP Standard Library Equivalent | Explanation |
| --- | --- | --- |
| `json.NewDecoder(r.Body).Decode(&body)` | `json_decode(file_get_contents('php://input'))` | Parsing incoming JSON payloads into native data structures. |
| `json.NewEncoder(w).Encode(response)` | `echo json_encode($response)` | Serializing structs/arrays into JSON for the HTTP response. |

## 6. Utilities & Environment

Basic server necessities like reading environment variables and logging errors.

| Go Implementation | PHP Standard Library Equivalent | Explanation |
| --- | --- | --- |
| `os.Getenv("DB_URL")` | `getenv('DB_URL')` || `$_ENV['DB_URL']` | Reading configuration from the environment. |
| `log.Printf(...)` | `error_log(...)` | Writing formatted outputs to the standard error stream. |
| `unicode.IsUpper(char)` | `ctype_upper($char)` || `preg_match('/[A-Z]/')` | Character inspection for the password strength validator. |

## 7. Email Delivery

Sending confirmation and password-reset emails over SMTP.

| Go Implementation (`net/smtp`) | PHP Standard Library Equivalent | Explanation |
| --- | --- | --- |
| `smtp.SendMail(...)` | `mail(...)` || raw SMTP via `fsockopen` | Sending outbound email over SMTP. |
| `smtp.PlainAuth(...)` | manual `AUTH PLAIN` SMTP command | SMTP authentication handshake. |

## 8. Image Processing & File Storage

Compositing uploaded photos with overlay PNGs and persisting them to disk.

| Go Implementation | PHP Standard Library Equivalent | Explanation |
| --- | --- | --- |
| `image.Decode` / `image/png` / `image/jpeg` | GD extension (`imagecreatefrompng`, `imagecreatefromjpeg`) | Decoding uploaded image bytes into an in-memory bitmap. |
| `image/draw` (`draw.Draw` with `draw.Over`) | GD's `imagecopy` / `imagealphablending` | Alpha-compositing the overlay PNG on top of the captured photo. |
| `image/png.Encode` | `imagepng(...)` | Encoding the composited bitmap back to PNG bytes for storage. |
| `mime/multipart.Reader` | `$_FILES` superglobal | Streaming/parsing an uploaded multipart file without a third-party form parser. |
| `os.WriteFile` / `os.MkdirAll` | `move_uploaded_file(...)` | Persisting the uploaded/composited file to local disk storage. |
| `github.com/google/uuid` (already a dependency) | `uniqid()` / `random_bytes()`-based filename | Generating collision-resistant random filenames instead of trusting client-supplied names. |