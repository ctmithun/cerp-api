# CERP API  
_A RESTful, AWS Lambda–ready backend for the Campus / Education Resource Planning (CERP) system, built in Go._

[![Go](https://img.shields.io/badge/Go-1.x-blue.svg)](https://go.dev/)
[![AWS Lambda](https://img.shields.io/badge/Deploy-AWS%20Lambda-orange.svg)](https://aws.amazon.com/lambda/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

---

## 🚀 Overview  
CERP-API is the backend system for managing academic workflows — admissions, attendance, student metadata, faculty records, and more.  
It’s built with **Go**, optimized for **AWS Lambda + API Gateway**, and supports **PostgreSQL / CockroachDB** for data and **AWS S3 / DynamoDB** for storage and metadata.

---

## 🧩 Tech Stack
| Component | Technology |
|------------|-------------|
| **Language** | Go (Golang) |
| **Cloud** | AWS Lambda + API Gateway |
| **Database** | PostgreSQL / CockroachDB |
| **Storage** | AWS S3 |
| **Metadata Store** | AWS DynamoDB |
| **Authentication** | JWT + HMAC Token |
| **Modules** | attendance, students, faculty, enquiry, uv, onboard_data, otp, notifications |

---

## ⚙️ Setup

### Prerequisites
- Go 1.20+  
- AWS credentials configured (`aws configure`)  
- Database (PostgreSQL/CockroachDB)  
- S3 bucket and DynamoDB tables set up

### Installation
```bash
git clone https://github.com/ctmithun/cerp-api.git
cd cerp-api
go mod download
```

### Configuration
Create your `.env` file (or config file in `cfg_details/`) with:
```env
AWS_REGION=ap-south-1
DB_URL=postgresql://user:pass@host:26257/cerp?sslmode=verify-full
JWT_SECRET=<jwt_secret>
S3_BUCKET=cerp-students
DYNAMO_METADATA_TABLE=college_metadata
DYNAMO_USER_TABLE=users_info
```

### Running Locally
```bash
go run main.go
```

### Deploy to AWS Lambda
```bash
GOOS=linux GOARCH=amd64 go build -o main
zip deployment.zip main
```
Upload `deployment.zip` to Lambda and connect via API Gateway.

---

## 🔐 Authentication Flow
1. Login → JWT issued (Auth header).  
2. `/roles` endpoint → Encoded HMAC token (`cerp-api-token`) returned.  
3. All subsequent requests → must include both JWT + CERP API token headers.

---

## 🧠 API Endpoints

### 🔸 Authentication & Roles
| Method | Endpoint | Description |
|--------|-----------|-------------|
| `GET` | `/roles` | Generate encoded CERP HMAC token |
| `GET` | `/userRoles` | Fetch user role list |

---

### 🔸 Admissions & Student Metadata
| Method | Endpoint | Description |
|--------|-----------|-------------|
| `POST` | `/admission/admit` | Admit a student with uploaded form data |
| `POST` | `/metadata/student/create` | Create new student metadata |
| `POST` | `/metadata/student/update` | Update student metadata |
| `DELETE` | `/metadata/student/delete` | Delete student metadata |
| `GET` | `/metadata/student/manage` | Manage all student metadata |
| `GET` | `/v2/metadata/student/manage` | Manage metadata (v2) |
| `PUT` | `/v2/metadata/student/update` | Update student metadata (v2) |
| `PUT` | `/metadata/student/usnUpdate` | Update USN/registration data |
| `POST` | `/metadata/student/upload` | Upload student documents |
| `GET` | `/metadata/student/files` | List student files |
| `DELETE` | `/metadata/student/files` | Delete a student file |
| `GET` | `/metadata/student/filedata` | Fetch metadata for student files |
| `GET` | `/metadata/student/vault/list` | Get vault metadata (multiple students) |
| `GET` | `/metadata/student/vault/student` | Get vault metadata (single student) |
| `POST` | `/metadata/student/vault/update` | Update student vault data |
| `POST` | `/metadata/student/vault/otp` | Generate OTP for vault operations |

---

### 🔸 Attendance
| Method | Endpoint | Description |
|--------|-----------|-------------|
| `GET` | `/attendance/students` | Fetch students for attendance |
| `POST` | `/attendance/update` | Update student attendance |
| `GET` | `/attendance/export` | Export attendance data |
| `GET` | `/subject` | Get subject list for attendance |

---

### 🔸 Faculty
| Method | Endpoint | Description |
|--------|-----------|-------------|
| `GET` | `/profile` | Get faculty profile |
| `POST` | `/profile` | Update faculty profile |
| `GET` | `/profile-photo` | Get faculty photo |
| `POST` | `/profile-photo` | Upload faculty photo |
| `GET` | `/metadata/faculty/manage` | Get faculty list |
| `POST` | `/metadata/faculty/create` | Create new faculty metadata |
| `POST` | `/metadata/faculty/update` | Update faculty metadata |
| `POST` | `/metadata/faculty/upload` | Upload faculty file |
| `GET` | `/metadata/faculty/files` | Get faculty files |
| `DELETE` | `/metadata/faculty/files` | Delete faculty file |
| `GET` | `/metadata/faculty/filedata` | Get faculty file metadata |
| `POST` | `/metadata/faculty/export` | Export faculty data |
| `DELETE` | `/metadata/faculty/delete` | Delete faculty profile |
| `DELETE` | `/metadata/faculty/deactivate` | Deactivate faculty |
| `POST` | `/metadata/faculty/profile` | Upload faculty photo |
| `GET` | `/metadata/faculty/profile` | Retrieve faculty photo |
| `GET` | `/metadata/faculty/subjects` | List assigned subjects |

---

### 🔸 Enquiry Management
| Method | Endpoint | Description |
|--------|-----------|-------------|
| `POST` | `/enq/create` | Create new enquiry |
| `PUT` | `/enq/update` | Update enquiry |
| `GET` | `/enq/list` | List enquiries |
| `GET` | `/enq/get` | Get enquiry by ID |
| `DELETE` | `/enq/delete` | Delete enquiry |
| `POST` | `/enq/comments/add` | Add comment |
| `GET` | `/enq/comments/get` | Get comments |

---

### 🔸 Metadata & Subject Mapping
| Method | Endpoint | Description |
|--------|-----------|-------------|
| `POST` | `/metadata/update` | Update general metadata |
| `GET` | `/metadata/fetch` | Fetch metadata |
| `GET` | `/metadata/s2s` | Fetch student-subject mapping |
| `GET` | `/metadata/getStudents` | Get student list |

---

### 🔸 UV (Document Validation)
| Method | Endpoint | Description |
|--------|-----------|-------------|
| `POST` | `/uv/create` | Create UV record |
| `GET` | `/uv/list` | List UV records |
| `POST` | `/uv/student/collect` | Collect student UV documents |

---

## 🧪 Testing
```bash
go test ./...
```

---

## 📁 Folder Structure
```
cerp-api/
├── attendance/
├── cfg_details/
├── enquiry/
├── faculty/
├── iam/
├── jwt/
├── notifications/
├── onboard_data/
├── otp/
├── psw_generator/
├── students/
├── subject/
├── u_by_service/
├── uv/
├── main.go
├── main_test.go
├── go.mod
└── .gitignore
```

---

## 👤 Author
**Mithun C Theertha**  
📧 [your.email@example.com](mailto:your.email@example.com)  
🔗 [https://github.com/ctmithun/cerp-api](https://github.com/ctmithun/cerp-api)

---

> _"Building smarter campuses, one API at a time."_ 🎓
