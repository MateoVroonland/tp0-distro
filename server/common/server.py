import socket
import logging
import signal

from common.utils import store_bets
from common.communication import CompleteSocket
from common.parser import parse_batch

MSG_TYPE_ACK="ACK"
MSG_TYPE_NACK="NACK"
MSG_TYPE_FIN="FIN"
MSG_TYPE_DRAW_READY="DRAW_READY"
MSG_TYPE_GET_WINNERS="GET_WINNERS"
MSG_TYPE_BATCH="BATCH"
MSG_TYPE_WINNERS="WINNERS"

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

    def process_batch(self, client_sock, bet_data):
        current_batch, errors = parse_batch(bet_data)
        
        if errors:
            logging.error(f"action: apuesta_recibida | result: fail | cantidad: {len(current_batch)}")
            client_sock.send_all(MSG_TYPE_NACK.encode('utf-8'), MSG_TYPE_NACK)
            return
        
        store_bets(current_batch)
        logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(current_batch)}')
        client_sock.send_all(MSG_TYPE_ACK.encode('utf-8'), MSG_TYPE_ACK)

    def __handle_client_connection(self, client_sock):
        try:
            bet_data, msg_type = client_sock.recv_all()
            while bet_data:
                if msg_type == MSG_TYPE_FIN:
                   break
                elif msg_type == MSG_TYPE_BATCH:
                    self.process_batch(client_sock, bet_data)
                elif msg_type == MSG_TYPE_DRAW_READY:
                    pass
                elif msg_type == MSG_TYPE_GET_WINNERS:
                    pass
                else:
                    logging.error(f"action: handle_client | result: fail | error: {e}")
                    client_sock.send_all(MSG_TYPE_NACK.encode('utf-8'), MSG_TYPE_NACK)
                    break
                bet_data, msg_type = client_sock.recv_all()
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
