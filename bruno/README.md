# Bruno API Collection - Campus Sensor Monitoring & Alerts

This directory contains a complete Bruno API collection covering all implemented HTTP routes for the **Campus Sensor Monitoring & Alerts** application.

## Prerequisites
1. Install [Bruno](https://www.usebruno.com/) (Open-source API Client).
2. Start the backend service:
   ```bash
   go run cmd/server/main.go
   ```

## Setup & Environment
1. Open Bruno and select **Open Collection**, then navigate to this `bruno/` directory.
2. Select the **Local** environment from the top-right environment selector.
3. The environment is pre-configured with:
   - `baseUrl`: `http://localhost:8080`
   - `adminEmail`: `organizer@campus.edu`
   - `adminPassword`: `CampusOrganizerSecure2026!`
   - `viewerEmail`: `viewer@campus.edu`
   - `viewerPassword`: `CampusViewerDemo2026!`
   - `sensorId`: `lib-fl2-temp`

## Recommended Run Order & Chaining
The requests are designed to run in natural sequence with automated variable chaining:

1. **Health & Readiness**
   - Run `Healthz Probe` -> expects 200 OK
   - Run `Readyz Probe` -> checks database readiness
2. **Authentication**
   - Run `Auth / Login Admin` -> authenticates organizer, validates signature, and automatically captures `authToken` into the environment.
   - Run `Auth / Get Current User` -> validates current claims with the captured Bearer token.
   - Run `Auth / Login Invalid` -> verifies 401 error response.
3. **Sensor Inventory & Ingestion**
   - Run `Sensors / List Sensors` -> lists all seeded campus sensors.
   - Run `Ingestion / Ingest Single Reading` -> submits single reading, verifies 202 Accepted.
   - Run `Ingestion / Ingest Batch Readings` -> submits burst batch, verifies 200 OK.
   - Run `Ingestion / Ingest Invalid Reading` -> verifies 400 validation error.
4. **Alert Rules Management**
   - Run `Alert Rules / Create Single Reading Rule` -> creates a rule and saves `ruleId` into the environment.
   - Run `Alert Rules / Create Time Window Rule` -> creates a 15-minute rolling average rule.
   - Run `Alert Rules / Get Rule by ID` -> retrieves the created rule.
   - Run `Alert Rules / Toggle Rule State` -> disables/enables the rule.
5. **Alerts & Lifecycle Actions**
   - Run `Simulator / Trigger Threshold Spike` -> simulates high heat exceeding the alert threshold.
   - Run `Alerts & Lifecycle / List Alerts` -> lists open incidents and captures `alertId`.
   - Run `Alerts & Lifecycle / Acknowledge Alert` -> marks incident as ACKNOWLEDGED with actor metadata.
   - Run `Alerts & Lifecycle / Resolve Alert` -> marks incident as RESOLVED with resolution note, clearing active deduplication key.
   - Run `Alerts & Lifecycle / Reopen Alert` -> reopens incident when needed.
6. **Analytics & Dashboard**
   - Run `Dashboard / Get Dashboard Summary` -> retrieves real-time sensor and alert aggregates.
   - Run `Readings & Trends / Get Metric Trends` -> retrieves time-bucketed aggregation points for charting.
   - Run `Dashboard / Get Audit Logs` -> lists audit trail of actions taken.
