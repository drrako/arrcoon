## Arrcoon

<p align="center">
  <img width="200" alt="image" src="https://github.com/user-attachments/assets/67b99312-e0c7-468b-8afa-f1b6737f6bbd" />
</p>

Arrcoon is a single-binary Go application that accompanies Sonarr or Radarr installation and tries its best to automatically remove associated torrents from the torrent client immediately when the media is deleted from the *arr library. It builds and maintains own series (movies) index based *arr on events and history API.

### Installation

Download the [latest](https://github.com/drrako/arrcoon/releases) binary for your architecture. Create a directory `arrcoon` inside *arr config folder and put the binary inside.
Make sure `arccoon` has execution rights:
```bash
mkdir /path/to/sonarr/config/arrcoon
cd /path/to/sonarr/config/arrcoon
curl -L "https://github.com/drrako/arrcoon/releases/download/65b2779/arrcoon-linux-amd64-65b2779.zip" -o arrcoon.zip && unzip arrcoon.zip && rm arrcoon.zip
chmod +x arrcoon
```

Create `config.yml` in the installation dir (`/path/to/sonarr/config/arrcoon`) based on the [example](https://github.com/drrako/arrcoon/blob/main/config.sample.yml), e.g.:
```yml
sonarr:
  host: http://localhost:8989
  token: XXXX
radarr:
  host: http://localhost:7878
  token: XXXX
clients:
  rtorrent:
    host: http://localhost/rtorrent/RPC2
```

Supported clients:
- rTorrent (basic auth is not supported)
  ```yml
  ...
  clients:
    rtorrent:
      host: http://localhost/rtorrent/RPC2
  ```
- transmission:
  ```yml
  ...
  clients:
    transmission:
      host: http://localhost:9091/transmission/rpc
  
  ... ---- Or with auth ---- ...
  
  clients:
    transmission:
      host: http://user:password@localhost:9091/transmission/rpc
  ```

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
