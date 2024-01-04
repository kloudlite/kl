## kl add mres

add mres to your kl-config file by selection from the all the mres available selected project

### Synopsis

```
Add env from managed resource

Using this command you are able to add a environment from the managed resource present on your project
Examples:
  # add managed resource by selecting one
  kl add mres

  # add managed resource providing resourceid and serviceid 
  kl add mres --resource=<resourceId> --service=<serviceId>

```

### Options

```
  -h, --help              help for mres
      --resource string   managed resource name
      --service string    managed service name
  -h, --help   help for mres
```

### SEE ALSO

* [kl add](kl_add.md)  - add [ secret | config | mres ] configuration to your kl-config file

###### Auto generated by kl CLI on 4-January-2024