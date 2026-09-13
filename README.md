# ![logo](.github/logo-small.png) Moreno.AlphaCore

---

## ❤️ Enjoy the Project or Want to Support?

[![AlphaCore Ko-Fi](https://www.ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/R6R21LO82)
[![Moreno.AlphaCore Go Port Ko-Fi](https://www.ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/denveous)

---

## Moreno.AlphaCore - Golang port of `alpha-core`

`Moreno.AlphaCore` is a Golang port of The Alpha Project's experimental emulator for version `0.5.3` of the **Friends & Family Alpha** of *World of Warcraft*.

This repository is the Go port of [The Alpha Project's Alpha Core](https://github.com/The-Alpha-Project/alpha-core). The upstream Python source remains the behavioral reference for this port.

> [!NOTE]
> The Go port is being developed in focused subsystems. The current slice provides SQLite data import, packet/SRP6 primitives, login, realm, proxy, world authentication, character create/list/delete, player-login and starting-inventory update packets, nearby creature and gameobject updates, item/page/quest/creature/gameobject queries, WHO, friend/ignore state, party invite/accept/disband/leader state, guild create/invite/roster/rank/chat state, channel join/list/chat/administration, played time, action and spell-button persistence, random roll, LFG, selection, target, chat/whisper, text emotes, readable item pages, basic inventory mutations, ping, time, movement persistence, and logout; broader gameplay remains in progress.

- [Database Tool](https://db.thealphaproject.eu/)
- [Original AlphaCore Discord](https://discord.gg/RzBMAKU)
- [MorenoLand Discord](https://discord.moreno.land)

---

## ⚙️ Configuration

The Go port uses SQLite and stores runtime state under `bin/` by default. The `--work` flag accepts either `--work=bin/` or `--work bin/` and points to the directory containing the SQLite files and built binary.

```bash
go run . --work=bin/
```

Bare `go run .` starts the translated listeners. Use `--bootstrap` when you only want to create or import the SQLite files and exit.

This creates `auth.sqlite3`, `realm.sqlite3`, `world.sqlite3`, and `dbc.sqlite3` under the selected work directory. The upstream SQL assets remain under `etc/databases` for later schema and data porting.

Create a local account with the legacy authentication path:

```bash
go run . --bootstrap --work=bin/ --create-account=PLAYER --password=PASSWORD
```

To load those retained SQL dumps into SQLite, run:

```bash
go run . --bootstrap --import-sql --work=bin/
```

---

## 📦 Installation

Install [Go](https://go.dev/dl/) and run from the repository root:

```bash
go run . --work=bin/
```

Build the executable into `bin/` with either `.\scripts\build.ps1` on Windows or `bash scripts/build.sh` on Unix-like systems. Run the built binary with `bin/Moreno.AlphaCore.exe --work=bin/` on Windows.

---

## 🖥 Client Setup

1. Create `realmlist.wtf` in the same folder as `WoW.exe`:
   ```
   SET realmlist "127.0.0.1"
   ```

2. (Optional) Clear the cache by adding the following before the `start` command in your batch file:
   ```
   Rmdir /S "WDB"
   ```

3. In-game, you may need to click **Change Realm** to log into your server.

4. On your first login, it is recommended to run:
   ```
   pwdchange
   ```
   Follow the on-screen instructions.

### Auth Options

#### Legacy
- Create `wow.ses`. The **first** and **second** lines must contain your `username` and `password`:
  ```
  username
  password
  ```
- To launch an **unmodified** client, start `WoWClient.exe` with the `-uptodate` parameter (and it's highly recommended to use `-windowed`). Example batch file:
  ```bat
  start WoWClient.exe -uptodate -windowed
  ```

#### SRP6
- **Login Server** — requires a `login.txt` file at the WoW root pointing to the login server, e.g.:
  ```
  127.0.0.1:3724
  ```
- **WoW.exe** should be executed with elevated admin rights so it can read/write `wow.ses`.
- **Update Server (Optional)** — requires an `Update.txt` file at `WoW/Data` pointing to the update server, e.g.:
  ```
  127.0.0.1:9081
  ```

---

## 🧩 Addons

[AlphaUI](https://github.com/The-Alpha-Project/AlphaUI) is a custom addon framework for the `0.5.3` client. It provides UI enhancements and quality-of-life features not present in the original client, communicating with the server through the addons chat API.

Addon protocol support will be ported with the corresponding client handlers.

---

## ⚠️ Common Issues

- **Work directory:** If the selected `--work` path is a file, choose a directory path instead.

- **SQLite files:** The Go bootstrap creates missing databases and preserves existing files. Remove the selected files manually when a clean local database is required.

> [!IMPORTANT]  
> Due to the age and experimental nature of the `0.5.3` client build, you may experience stability and performance issues. These are client-related and **not** caused by the core server implementation.

---

## Disclaimer

The `Alpha Project` **does not** distribute a client. You will need to obtain a clean `0.5.3` client yourself.

The `Alpha Project` **does not** encourage unofficial public servers. If you use this project to run an unofficial public server rather than for testing and learning, it is your personal choice.

---

## License

The original `Alpha Project` source components retained in this repository, and the Moreno.AlphaCore port, are released under the [GPL-3.0](https://www.gnu.org/licenses/gpl-3.0.en.html) license.

`Moreno.AlphaCore` is **not** an official Blizzard Entertainment product and is **not** affiliated with or endorsed by *World of Warcraft* or Blizzard Entertainment.
