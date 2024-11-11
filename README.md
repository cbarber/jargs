## jargs

Map json stdin to a xargs style command for json stdout.

This will retrieve all secrets with decoded values as an object
```
aws secretsmanager list-secrets |
    jq "[ .SecretList[].Name ]" | 
    jargs -map "{ {{.In|tojson}}: {{.Out.SecretString}} }" aws secretsmanager get-secret-value --secret-id '{{.In}}' |
    jq .
```


## TODO

* Support scanf lines style in stdin and stdout. 
* Support parallel execution
* Support iterating key/value pairs for an object
* CI for tests and release
