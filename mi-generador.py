
import sys
import random
from datetime import datetime, timedelta

def generate_random_name():
    first_names = ["Enzo", "Emiliano", "Julian", "Lionel", "Mateo"]
    return random.choice(first_names)

def generate_random_surname():
    surnames = ["Messi", "Rodriguez", "Martinez", "Alvarez", "Fernandez"]
    return random.choice(surnames)

def generate_random_document():
    return str(random.randint(10000000, 55000000))

def generate_random_birthdate():
    start_date = datetime(1980, 1, 1)
    end_date = datetime(2024, 12, 31)
    
    delta_days = (end_date - start_date).days
    random_days = random.randint(0, delta_days)
    random_date = start_date + timedelta(days=random_days)
    
    return random_date.strftime("%Y-%m-%d")

def generate_random_bet_number():
    return str(random.randint(0, 9999)).zfill(4)

def generate_compose(output_file, num_clients):
    compose_content = f"""name: tp0
services:
  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    environment:
      - NUM_AGENCIES={num_clients}
    networks:
      - testing_net
    volumes:
      - ./server/config.ini:/config.ini
"""

    for i in range(1, num_clients + 1):
        nombre = generate_random_name()
        apellido = generate_random_surname()
        documento = generate_random_document()
        nacimiento = generate_random_birthdate()
        numero = generate_random_bet_number()
        client_section = f"""  client{i}:
    container_name: client{i}
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID={i}
      - NOMBRE="{nombre}"
      - APELLIDO="{apellido}"
      - DOCUMENTO={documento}
      - NACIMIENTO={nacimiento}
      - NUMERO={numero}
    networks:
      - testing_net
    depends_on:
      - server
    volumes:
      - ./client/config.yaml:/config.yaml
      - ./.data/agency-{i}.csv:/agency.csv
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