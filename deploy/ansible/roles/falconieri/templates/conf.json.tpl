{
  "providers": {
    "snom": {
      "user":"{{ snom_user }}",
      "password": "{{ snom_password }}",
      "rpc_url": "{{ snom_rpc_url }}",
      "disable": {{ snom_disable }}
    },
    "yealink": {
      "user":"{{ yealink_user }}",
      "password": "{{ yealink_password }}",
      "rpc_url": "{{ yealink_rpc_url }}",
      "disable": {{ yealink_disable }}
    },
    "giagaset": {
      "user":"{{ gigaset_user }}",
      "password": "{{ gigaset_password }}",
      "rpc_url": "{{ gigaset_rpc_url }}",
      "disable": {{ gigaset_disable }},
      "disable_crc": {{ gigaset_disable_crc }}
    },
    "fanvil": {
      "user":"{{ fanvil_user }}",
      "password": "{{ fanvil_password }}",
      "rpc_url": "{{ fanvil_rpc_url }}",
      "disable": {{ fanvil_disable }}
    }
  }
}