
import sys

def generate_compose(output_file, num_clients):
    compose_content = f"""name: tp0
services:
  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
    networks:
      - testing_net
    volumes:
      - ./server/config.ini:/config.ini
"""

    for i in range(1, num_clients + 1):
        client_section = f"""  client{i}:
    container_name: client{i}
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID={i}
    networks:
      - testing_net
    depends_on:
      - server
    volumes:
      - ./client/config.yaml:/config.yaml
"""
        compose_content += client_section

    compose_content += """networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
"""

    with open(output_file, 'w') as f:
        f.write(compose_content)


if __name__ == "__main__":
    if len(sys.argv) != 3:
        print(f"Uso: {sys.argv[0]} <nombre-archivo-salida> <cantidad-clientes>")
        sys.exit(1)
    
    output_file = sys.argv[1]
    num_clients = int(sys.argv[2])
    if num_clients < 0:
        print("Error: La cantidad de clientes no debe ser negativa") 
        sys.exit(1)
    
    generate_compose(output_file, num_clients)