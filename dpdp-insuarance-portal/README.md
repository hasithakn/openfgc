# Internal Data Sharing — Life Insurance Consent Demo

A self-contained demo showing how a life insurance company collects, manages, and enforces user consent for personal data sharing using **OpenFGC** (Open Fine-Grained Consent).

---

## What This Demo Shows

A customer visits a life insurance website and requests a quotation. Before submitting their details, they are shown **exactly why each piece of data is collected** and given a clear, granular choice over optional data use (email marketing). On submission, a structured consent record is created in the Consent Manager — capturing what was agreed to, by whom, and for what purpose.

This demonstrates:
- **Transparency** — purpose descriptions are fetched live from the Consent Manager and displayed inline in the form
- **Granular consent** — mandatory data collection (policy creation) is separated from optional marketing consent
- **Consent as a record** — every submission creates a verifiable, auditable consent entry in OpenFGC with a unique ID
- **Data minimisation** — only the fields required for each declared purpose are collected

---

## Data Elements

| Element | Display Name | Used For |
|---|---|---|
| `name` | Full Name | Policy personalisation |
| `email` | Email Address | Quote delivery, policy documents |
| `age` | Age | Premium calculation |
| `marketing_via_email` | Agree to marketing via email | Optional marketing consent flag |

---

## Consent Purposes

| Purpose | Elements | Type |
|---|---|---|
| `create_custom_insurance_policy` | `name`, `email`, `age` | Mandatory (required for quotation) |
| `marketing_via_email` | `marketing_via_email` | Optional (user opt-in) |

---

## Demo Flow

```
1. Customer lands on the Life Insurance home page
       ↓
2. Clicks "Get Your Free Quotation"
       ↓
3. Quotation form loads — purpose descriptions are fetched
   live from the Consent Manager and shown inline
       ↓
4. Customer fills in: Name, Email, Age
   • Required: agrees to Privacy Policy (creates_custom_insurance_policy)
   • Optional: opts in to email marketing (marketing_via_email)
       ↓
5. On submit → consent record created in OpenFGC (DEMO-ORG-002)
   with authorisation entry containing the user's PII
       ↓
6. Thank You page shows the Consent Record ID from OpenFGC
```

---

## Setup

**Prerequisites:** OpenFGC running on `http://127.0.0.1:3000`

### 1. Install & start
```bash
cd demo-ui-internal-data-share-usecase
npm install
node server.js
```
Runs on **http://localhost:3020**

### 2. One-time consent configuration
Open **http://localhost:3020/setup/** and click **Create in Consent Manager**.

This registers the 4 elements and 2 purposes above in OpenFGC under org `DEMO-ORG-002`. Safe to run multiple times — existing records are skipped.

### 3. Run the demo
Open **http://localhost:3020** — the home page is ready.

---

## Configuration

Edit `config.json` to change the company name before a demo:

```json
{
  "companyName": "ABC Life Insurance Provider",
  "orgId": "DEMO-ORG-002",
  "openfgcUrl": "http://127.0.0.1:3000",
  "port": 3020
}
```

---

## Pages

| URL | Description |
|---|---|
| `/` | Life insurance home page |
| `/quotation.html` | Quotation form with live consent purpose display |
| `/thank-you.html` | Submission confirmation with Consent Record ID |
| `/setup/` | One-time consent elements & purposes configuration |
