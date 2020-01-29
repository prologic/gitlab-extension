**To deploy application**

```
$ cd <project_dir>

$ docker volume create gitlab_extension_db

$ docker build -t gitlab-extension .

$ docker run -d -p port:port \
             --name name --restart unless-stopped \ 
             -v config.yaml:/app/config.yaml \ 
             -v gitlab_extensions_db:/app/db
```