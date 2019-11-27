# 新建索引

```
PUT 172.16.13.127:9200/majorana

{
  "mappings": {
    "jobs": {
      "include_in_all": false,
      "properties": {
        "first_run_at": {
          "type": "long"
        },
        "job_type": {
          "type": "keyword"
        },
        "core_time": {
          "type": "long"
        },
        "doc_update_time": {
          "type": "long"
        },
        "groups": {
          "type": "keyword"
        },
        "cluster": {
          "type": "keyword"
        },
        "cores": {
          "type": "long"
        },
        "error_times": {
          "type": "long"
        },
        "first_create_at": {
          "type": "long"
        },
        "group": {
          "type": "keyword"
        },
        "handle": {
          "type": "keyword"
        },
        "image": {
          "type": "keyword"
        },
        "input": {
          "type": "keyword"
        },
        "last_output": {
          "type": "keyword"
        },
        "last_run_status": {
          "type": "long"
        },
        "last_update_at": {
          "type": "long"
        },
        "long_time": {
          "type": "long"
        },
        "memory": {
          "type": "long"
        },
        "n_cores": {
          "type": "long"
        },
        "name": {
          "type": "text"
        },
        "options": {
          "type": "text",
          "index": "no"
        },
        "pause": {
          "type": "boolean"
        },
        "remove": {
          "type": "boolean"
        },
        "rerun": {
          "type": "boolean"
        },
        "run_detail": {
          "type": "text",
          "index": "no"
        },
        "run_times": {
          "type": "long"
        },
        "sys_prio": {
          "type": "float"
        },
        "timeout": {
          "type": "long"
        },
        "total_run_time": {
          "type": "long"
        },
        "try_times": {
          "type": "long"
        },
        "user": {
          "type": "keyword"
        },
        "user_prio": {
          "type": "float"
        }
      }
    }
  }
}
```

# 更新索引

```
PUT /majorana/_mapping/jobs
{
	"properties": {
		"first_run_at": {
			"type": "long"
		},
		"job_type": {
			"type": "keyword"
		},
		"core_time": {
			"type": "long"
		},
		"doc_update_time": {
			"type": "long"
		},
		"groups": {
			"type": "keyword"
		},
		"cluster": {
			"type": "keyword"
		},
		"cores": {
			"type": "long"
		},
		"error_times": {
			"type": "long"
		},
		"first_create_at": {
			"type": "long"
		},
		"group": {
			"type": "keyword"
		},
		"handle": {
			"type": "keyword"
		},
		"image": {
			"type": "keyword"
		},
		"input": {
			"type": "keyword"
		},
		"last_output": {
			"type": "keyword"
		},
		"last_run_status": {
			"type": "long"
		},
		"last_update_at": {
			"type": "long"
		},
		"long_time": {
			"type": "long"
		},
		"memory": {
			"type": "long"
		},
		"n_cores": {
			"type": "long"
		},
		"name": {
			"type": "text"
		},
		"options": {
			"type": "text",
			"index": "no"
		},
		"pause": {
			"type": "boolean"
		},
		"remove": {
			"type": "boolean"
		},
		"rerun": {
			"type": "boolean"
		},
		"run_detail": {
			"type": "text",
			"index": "no"
		},
		"run_times": {
			"type": "long"
		},
		"sys_prio": {
			"type": "float"
		},
		"timeout": {
			"type": "long"
		},
		"total_run_time": {
			"type": "long"
		},
		"try_times": {
			"type": "long"
		},
		"user": {
			"type": "keyword"
		},
		"user_prio": {
			"type": "float"
		}
	}
}
```

# 修改限制

```
PUT 127.0.0.1:8888/_cluster/settings
{
    "persistent" : {
        "script.max_compilations_per_minute" : 500
    }
}
```