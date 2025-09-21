## Arrcoon

<p align="center">
  <img width="200" alt="arrcoon-logo" src="https://github.com/user-attachments/assets/cae8cf05-ed86-411b-94c9-0a6505585204" />
</p>

Arrcoon is a single-binary Go application that accompanies Sonarr or Radarr installation and tries its best to automatically remove associated torrents from the torrent client immediately when the media is deleted from the *arr library. It builds and maintains own series (movies) index based *arr on events and history API.

### Key features

- Instantly removes torrent downloads from the client when the corresponding show/movie is removed from *arrs library
- Removes torrent downloads that are no longer mapped in the *arrs library (including individual episodes/season packs)

### Installation

Download the [latest](https://github.com/drrako/arrcoon/releases) binary for your architecture. Create a directory `arrcoon` inside *arr config folder and put the binary inside.
Make sure `arccoon` has execute permissions:
```bash
mkdir /path/to/sonarr/config/arrcoon
cd /path/to/sonarr/config/arrcoon
curl -L "https://github.com/drrako/arrcoon/releases/download/2025.9.0/arrcoon-linux-amd64-2025.9.0.zip" -o arrcoon.zip && unzip arrcoon.zip && rm arrcoon.zip
chmod +x arrcoon
```

Create `config.yml` in the installation dir (`/path/to/sonarr/config/arrcoon`) based on the [example](https://github.com/drrako/arrcoon/blob/main/config.sample.yml), e.g.:
```yml
sonarr:
  host: http://localhost:8989
  token: XXXX
# radarr:
#   host: http://localhost:7878
#   token: XXXX
clients:
  rtorrent:
    host: http://localhost/rtorrent/RPC2
```

> ⚠️ arrcoon supports only single torrent client per *arr installation

| Client | Configuration Example | Notes |
| :--- | :--- | :--- |
| **rTorrent** | `clients:`<br>`  rtorrent:`<br>`    host: http://localhost/rtorrent/RPC2` | Basic auth is not supported. |
| **transmission** | `# Without auth<br>clients:`<br>`  transmission:`<br>`    host: http://localhost:9091/transmission/rpc`<br>`# With auth`<br>`clients:`<br>`  transmission:`<br>`    host: http://user:password@localhost:9091/transmission/rpc` | |
| **qbittorrent** | `# Without auth`<br>`clients:`<br>`  qbittorrent:`<br>`    host: http://localhost:8080`<br>`# With auth`<br>`clients:`<br>`  qbittorrent:`<br>`    host: http://LOGIN:PASS@localhost:8080` | |


Add Sonarr/Radarr `arrcoon` connection and click `Test` to validate config:

<p align="center">
  <img width="300" alt="sonarr1" src="https://github.com/user-attachments/assets/795424ae-363d-44bb-9cbf-8f95b7877d58" />
  <img width="243" alt="sonarr2" src="https://github.com/user-attachments/assets/883724d2-6b2e-4a27-85ae-3f95cc65443a" />
</p>

> :warning: You're required to click `Test` as arrcoon builds internal index during testing


### Logs

Logs can be found in the `logs` directory, alongside the `arrcoon` binary:
```bash
tail -f /path/to/sonarr/config/arrcoon/logs/arrcoon.log
```
