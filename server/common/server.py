import socket
import logging
import signal

from common.utils import store_bets
from common.communication import CompleteSocket
from common.parser import parse_batch

ACK_MESSAGE= "ACK"
NACK_MESSAGE= "NACK"
FIN_MESSAGE= "FIN"

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        signal.signal(signal.SIGTERM, self.__signal_handler)
        self._running = True
        self._active_client_connection = None

    def __signal_handler(self, signum, frame):
        self._running = False
        self._active_client_connection.close()
        self._server_socket.close()

    def send_ack(self, client_sock):
        ack = ACK_MESSAGE
        ack_bytes = ack.encode('utf-8')
        message = f"{len(ack_bytes)}:".encode('utf-8') + ack_bytes
        client_sock.send_all(message)

    def send_nack(self, client_sock):
        nack = NACK_MESSAGE
        nack_bytes = nack.encode('utf-8')
        message = f"{len(nack_bytes)}:".encode('utf-8') + nack_bytes
        client_sock.send_all(message)

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        while self._running:
            client_sock = CompleteSocket(self.__accept_new_connection())
            if client_sock is not None:
                self._active_client_connection = client_sock
                self.__handle_client_connection(client_sock)

    def __handle_client_connection(self, client_sock):
        try:
            bet_data = client_sock.recv_all()
            current_batch = []
            while bet_data:
                if bet_data == FIN_MESSAGE:
                    logging.info(f"action: fin_received | result: success")
                    break
                current_batch, errors = parse_batch(bet_data)
                
                if errors:
                    logging.error(f"action: apuesta_recibida | result: fail | cantidad: {len(current_batch)}")
                    self.send_nack(client_sock)
                    return
                
                store_bets(current_batch)
                logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(current_batch)}')
                self.send_ack(client_sock)
                
                bet_data = client_sock.recv_all()
        except Exception as e:
            logging.error(f"action: handle_client | result: fail | error: {e}")
        finally:
            client_sock.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        try:
            logging.info('action: accept_connections | result: in_progress')
            c, addr = self._server_socket.accept()
            logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
            return c
        except OSError as e:
            return None
