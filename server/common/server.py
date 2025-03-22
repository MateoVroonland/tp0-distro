import socket
import logging
import signal

from common.utils import Bet, store_bets


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
        if self._active_client_connection:
            self._active_client_connection.shutdown(socket.SHUT_RDWR)
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
            client_sock = self.__accept_new_connection()
            if client_sock is not None:
                self._active_client_connection = client_sock
                self.__handle_client_connection(client_sock)

    def send_all(self, sock, data):
        total_sent = 0
        while total_sent < len(data):
            sent = sock.send(data[total_sent:])
            if sent == 0:
                raise OSError
            total_sent += sent

    def recv_length_prefix(self, sock):
        prefix_buffer = bytearray()
        
        while b':' not in prefix_buffer:
            chunk = sock.recv(1)
            if not chunk:
                raise ConnectionError("Connection closed while reading length prefix")
            prefix_buffer.extend(chunk)
        
        length_str = prefix_buffer.split(b':')[0].decode('utf-8')
        try:
            return int(length_str)
        except ValueError:
            raise ValueError(f"Invalid length prefix: {length_str}")

    def recv_exact(self, sock, length):
        payload = bytearray()
        remaining = length
        
        while remaining > 0:
            chunk = sock.recv(min(remaining, 1024))
            
            if not chunk:
                raise ConnectionError(f"Connection closed after reading {length - remaining} of {length} bytes")
            
            payload.extend(chunk)
            remaining -= len(chunk)
        
        return bytes(payload)

    def recv_all(self, sock):
        try:
            length = self.recv_length_prefix(sock)
            payload = self.recv_exact(sock, length)
            return payload.decode('utf-8')
        except (ConnectionError, ValueError) as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
            raise


    def __handle_client_connection(self, client_sock):
        try:
            bet_data = self.recv_all(client_sock)
            fields = bet_data.rstrip('\n').split(',')
            
            if len(fields) != 5:
                error_msg = f"Invalid number of fields: {len(fields)}"
                logging.error(f"action: receive_bet | result: fail | error: {error_msg}")
                return
            
            name, surname, document_id, birth_date, number = fields
            
            addr = client_sock.getpeername()
            logging.info(f'action: receive_bet | result: success | ip: {addr[0]} | dni: {document_id} | numero: {number}')
            bet = Bet(agency=1, first_name=name, last_name=surname, document=document_id, birthdate=birth_date, number=number)

            store_bets([bet])
            logging.info(f'action: apuesta_almacenada | result: success | dni: {document_id} | numero: {number}')
        
            confirmation = "ok\n"
            confirmation_encoded = f"{len(confirmation)}:{confirmation}".encode('utf-8')
            self.send_all(client_sock, confirmation_encoded)    
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
