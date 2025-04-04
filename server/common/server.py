import os
import socket
import logging
import signal
import multiprocessing as mp
from multiprocessing import Manager, Lock, Barrier

from common.utils import store_bets, winners_for_agency
from common.communication import CompleteSocket
from common.parser import parse_batch

MSG_TYPE_ACK="ACK"
MSG_TYPE_NACK="NACK"
MSG_TYPE_FIN="FIN"
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
        self._expected_agencies = int(os.getenv('NUM_AGENCIES', 5))

        self._manager = Manager()
        self._file_lock = Lock()
        self._running = self._manager.Value('b', True)
        self._barrier = Barrier(self._expected_agencies)
        self._processes = []

    def shutdown(self):
        self._running.value = False
        for proc in self._processes:
            proc.join()
        self._server_socket.close()

    def __signal_handler(self, signum, frame):
        logging.info(f"action: signal_handler | result: success | signal: {signum}")
        self.shutdown()

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        while self._running.value:
            client_sock = self.__accept_new_connection()
            if client_sock is not None:
                complete_sock = CompleteSocket(client_sock)
                proc = mp.Process(target=self.__handle_client_connection, args=(complete_sock,))
                self._processes.append(proc)
                proc.start()

    def process_batch(self, client_sock, bet_data):
        current_batch, errors = parse_batch(bet_data)
        agency = current_batch[0].agency
        
        if errors:
            logging.error(f"action: apuesta_recibida | result: fail | cantidad: {len(current_batch)}")
            client_sock.send_all(MSG_TYPE_NACK.encode('utf-8'), MSG_TYPE_NACK)
            return
        
        with self._file_lock:
            store_bets(current_batch)
        
        logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(current_batch)}')
        client_sock.send_all(MSG_TYPE_ACK.encode('utf-8'), MSG_TYPE_ACK)
        return agency

    def process_draw(self, agency_id):
        with self._file_lock:
            winners = winners_for_agency(agency_id)
        return winners
    
    def send_winners(self, client_sock, winners):
        winners_joined = ",".join(winners)
        winners_msg = winners_joined + "\n"
        client_sock.send_all(winners_msg.encode('utf-8'), MSG_TYPE_WINNERS)

    def __handle_client_connection(self, client_sock):
        try:
            agency_id = 0
            bet_data, msg_type = client_sock.recv_all()
            while bet_data and self._running.value:
                if msg_type == MSG_TYPE_FIN:
                    logging.info(f"action: send batches finished | result: success")
                elif msg_type == MSG_TYPE_BATCH:
                    agency_id = self.process_batch(client_sock, bet_data)
                elif msg_type == MSG_TYPE_GET_WINNERS:
                    self._barrier.wait()
                    agency_id = int(bet_data.strip())
                    logging.info(f"action: sorteo | result: success")
                    winners = self.process_draw(agency_id)
                    self.send_winners(client_sock, winners)
                    break
                else:
                    logging.error(f"action: handle_client | result: fail | error: Invalid message type")
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
