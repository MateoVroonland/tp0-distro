from common.utils import Bet

def parse_batch(payload):
    bets = []
    errors = 0
    for line in payload.split("\n"):
        if line:
            try:
                agency, nombre, apellido, documento, nacimiento, numero = line.split(",")
                bet = Bet(agency, nombre, apellido, documento, nacimiento, numero)
                bets.append(bet)
            except Exception as e:
                errors += 1
    return bets, errors
    