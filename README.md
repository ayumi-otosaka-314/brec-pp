# brec-pp

Post processors for [Bililive Recorder](https://github.com/BililiveRecorder/BililiveRecorder). 

---

## Major features 
- [x] Auto remove recordings from local file system to ensure capacity for incoming recording, starting from the oldest. 
- [x] Auto upload recorded archive to **Google Drive** when recording file is completed. 
  - [x] Auto remove recordings from google drive to ensure capacity for upload, starting from the oldest. 
- [x] Send notification to **Discord** via [Webhook](https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks) on below events: 
  - Recording started
  - Recording finished, file ready to be uploaded 
  - Upload finished 

---
## Installation
* **Executables**   
  Executables will be uploaded to [release page](https://github.com/ayumi-otosaka-314/brec-pp/releases) in future releases. 
* **Build from source**   
  Ensure that [golang is installed](https://golang.org/doc/install) on your system.
  ```Bash
  git clone https://github.com/ayumi-otosaka-314/brec-pp && cd brec-pp
  go build -o ./bin/
  # Executable: ./bin/brec-pp
  ```

---
## Usage 
### Integration with [Bililive Recorder](https://github.com/BililiveRecorder/BililiveRecorder)
Paste the configured URL under `Webhook V2` section of Bililive Recorder.  
Also, please configure storage `rootPath` to be the same as the working directory of Bililive Recorder. 

### Configuration 
Example of configuration file could be found at [config/example.yaml](config/example.yaml).  
The path of the configuration file should be passed explicitly to the executable via `--config` arg. Example: 
```bash
./bin/brec-pp --config ./config/example.yaml
```
Google Drive and Discord notification could be configured for individual streamers by RoomID. Otherwise, it will fallback to default configuration. 

### Google Drive
#### Authentication 
To upload to google drive, this application have to be authenticated via some JSON credentials.  
Please create a service account with a JSON key and download it. Please refer to [this article](https://cloud.google.com/iam/docs/creating-managing-service-account-keys) for more details.  
- [ ] Potential feature: adding support for personal account is possible. 

#### Folder ID 
Please refer to [this guide](https://robindirksen.com/blog/where-do-i-get-google-drive-folder-id) for more details. 

### Discord 
Please refer to [this guide](https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks) to create a webhook, and paste the URL in the configuration file. 
