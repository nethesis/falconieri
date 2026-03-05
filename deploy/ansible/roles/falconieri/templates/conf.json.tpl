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
    "gigaset": {
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
    },
    "ymcs": {
      "base_url": "{{ ymcs_base_url }}",
      "client_id": "{{ ymcs_client_id }}",
      "client_secret": "{{ ymcs_client_secret }}",
      "disable": {{ ymcs_disable }}
    },
    "grape": {
      "base_url": "{{ grape_base_url }}",
      "client_id": "{{ grape_client_id }}",
      "client_secret": "{{ grape_client_secret }}",
      "disable": {{ grape_disable }}
    },
    "sraps": {
      "base_url": "{{ sraps_base_url }}",
      "client_id": "{{ sraps_client_id }}",
      "client_secret": "{{ sraps_client_secret }}",
      "disable": {{ sraps_disable }}
    }
  }
}