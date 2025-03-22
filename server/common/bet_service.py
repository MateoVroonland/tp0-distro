from common.utils import Bet

class BetService:
    # format: agency,nombre,apellido,documento,nacimiento,numero\n
    @staticmethod
    def parse_bet(payload):
        try:
            agency, nombre, apellido, documento, nacimiento, numero = payload.rstrip('\n').split(',')
            return Bet(agency, nombre, apellido, documento, nacimiento, numero)
        except ValueError as e:
            raise ValueError(f"Invalid bet payload: {payload}")

    