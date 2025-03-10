## Arrcoon

<p align="center">
  <img width="200" alt="image" src="https://github.com/user-attachments/assets/67b99312-e0c7-468b-8afa-f1b6737f6bbd" />
</p>

Arcoon is a single-binary Go application compiled for various platforms that accompanies Sonarr or Radarr installation and tries its best to automatically remove associated torrents from the torrent client as soon as the media is deleted from the *arr library. It builds and maintains own series (movies) index based *arr on events and history API.

### Installation

Download the [latest](https://github.com/drrako/arrcoon/releases) appropriate architecture binary. Create a directory `arrcoon` inside your *arr config folder and put there the binary.
Make sure `arccoon` has execution rights:
```bash
cd /path/to/sonarr/config/arrcoon
chmod +x arrcoon
```

Add `arrcoon` to Sonarr/Radarr:

<p align="center">
  <img width="300" alt="sonarr1" src="https://github.com/user-attachments/assets/795424ae-363d-44bb-9cbf-8f95b7877d58" />
  <img width="243" alt="sonarr2" src="https://github.com/user-attachments/assets/883724d2-6b2e-4a27-85ae-3f95cc65443a" />
</p>

> :warning: You're required to click `Tests` as it builds internal index if successful.
