import socket

PARAMETER_DELIMITER = b':'

class CompleteSocket:
    def __init__(self, socket):
        self._sock = socket
        self._buffer = bytearray()

    def append_message_type(self, message_type, message):
        type_bytes = message_type.encode('utf-8')
        return type_bytes + PARAMETER_DELIMITER + message

    def append_message_length(self, message):
        length_bytes = str(len(message)).encode('utf-8')
        return length_bytes + PARAMETER_DELIMITER + message

    def send_all(self, data, message_type):
        typed_data = self.append_message_type(message_type, data)
        final_data = self.append_message_length(typed_data)
        total_sent = 0
        while total_sent < len(final_data):
            sent = self._sock.send(final_data[total_sent:])
            if sent == 0:
                raise OSError
            total_sent += sent
    
    def recv_length_prefix(self): 
        self._buffer.clear()       
        while PARAMETER_DELIMITER not in self._buffer:
            chunk = self._sock.recv(1)
            if not chunk:
                raise ConnectionError("Connection closed while reading length prefix")
            self._buffer.extend(chunk)
        
        length_str = self._buffer.split(PARAMETER_DELIMITER)[0].decode('utf-8')
        try:
            return int(length_str)
        except ValueError:
            raise ValueError(f"Invalid length prefix: {length_str}")

    def recv_exact(self, length):
        self._buffer.clear()
        remaining = length
        
        while remaining > 0:
            chunk = self._sock.recv(min(remaining, 1024))
            
            if not chunk:
                raise ConnectionError(f"Connection closed after reading {length - remaining} of {length} bytes")
            
            self._buffer.extend(chunk)
            remaining -= len(chunk)
        
        return bytes(self._buffer)

    def recv_all(self):
        try:
            length = self.recv_length_prefix()
            fields = self.recv_exact(length).split(PARAMETER_DELIMITER)
            message_type = fields[0].decode('utf-8')
            payload = fields[1].decode('utf-8')
            return payload, message_type
        except (ConnectionError, ValueError) as e:
            raise

    def getpeername(self):
        return self._sock.getpeername()

    def close(self):
        if self._sock is not None:
            try:
                self._sock.shutdown(socket.SHUT_RDWR)
            except OSError:
                pass
            self._sock.close()

        