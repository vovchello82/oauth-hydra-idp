# hydra-id-provider

Die App spielt für den Hydra oAuth Broker die Rolle eines identity Providers and stellt die Login bzw. Consent Logik zur Verfügung. 

## How To Use

hydra-id-provider übernimmt die Registrierung der Clients beim oAuth Broker, falls eine JSON Datei namens clients.json in dem /import exisitert.
Die Datei kann einfach mit der docker-compose Deklaration ` volumes - [HOST_PATH_TO_JSON]:/import/clients.json` in den Container eingebunden werden. Eine Beispile Datei ist in `/import/clients.json` verfügbar.
Nach demselben Prinzip lassen sich auch Cresdentials für Benutzer einbinden. Eine Beispieldatei ist in `/import/users.json` verfügbar.

### ENVS

 - **HYDRA_ADMIN_URL** *Required* Der Hydra Admin Endpoint
 - **HYDRA_PUBLIC_URL**  *Required* Der Hydra Public Endpoint
 - **ALTERNATIVE_REDIRECT_HYDRA_URL** *Optional* Überschreibt die standard redirect URL 
 - **ISSUER_URI** *Optional* Falls der Issuer ein anderer als **HYDRA_PUBLIC_URL** ist

### HTTPS, TLS/SSL Certificates

Beim Starten generiert Hydra ein self-signed Zertifikat, welches für HTTPS Verbindungen verwenden werden. Der jewelige Klient soll diesem Zertifikat vertrauen oder die TLS-Verifizierung deaktivieren, um mit Hydra zu kommunizieren.
Hydra kann auch ein bestimmtes Zertifikat für die HTTPS-Verbindungen verwenden. 

Falls man mit unsichiren HTTP Verbinden fortsetzen möchte, kann es per Flag `-dangerous-force-http` aktivieren.

### Docker

* hydra 
* hydra-id-provider


### Links
* https://auth0.com/docs/get-started/authentication-and-authorization-flow/authorization-code-flow
* https://www.ory.sh/docs/hydra/install#docker
* https://www.ory.sh/docs/hydra/reference/api

