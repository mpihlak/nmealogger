#!/usr/bin/env python3

import datetime
import logging
import os
import socket
import time

HOST = '127.0.0.1'
PORT = 10110
OUTPUT_DIRECTORY = "data"
RETRY_INTERVAL_SECONDS = 1
REPORTING_INTERVAL_SECONDS = 60
ROTATION_INTERVAL_SECONDS = 3600

def valid_nmea_checksum(sentence):
    pieces = sentence.split('*')
    if len(pieces) != 2:
        return False

    data, provided_checksum = pieces
    data = data.lstrip('$')

    checksum = 0
    for char in data:
        checksum ^= ord(char)

    return hex(checksum)[2:].upper().zfill(2) == provided_checksum

def read_and_log(s):
    while True:
        file_ts = datetime.datetime.now().strftime('%Y-%m-%dT%H:%M:%S')
        filename = os.path.join(OUTPUT_DIRECTORY, f"nmea-{file_ts}.log")
        logging.info(f"Writing to {filename}")

        ok_sentences = 0
        nok_sentences = 0
        last_reported_time = time.time()
        last_rotation_time = time.time()
        with open(filename, 'wb') as f:
            while True:
                # Note: deliberately ignoring possible newline boundary issues as these are
                # very rare due to the low data rates.
                data = s.recv(4096)
                if not data:
                    logging.info("Connection closed.")
                    return
                
                strdata = data.decode('ascii', errors='ignore')
                if not strdata.endswith('\n'):
                    logging.warning(f"Data does not end in newline: {strdata}")
                else:
                    strdata = strdata.rstrip()
                sentences = strdata.split('\n')

                ts = datetime.datetime.now(datetime.timezone.utc).isoformat()
                for sentence in sentences:
                    if valid_nmea_checksum(sentence):
                        f.write(f"{ts}\t{sentence}\n".encode('ascii'))
                        f.flush()
                        ok_sentences += 1
                    else:
                        logging.warning(f"Skipping sentence with invalid checksum: {sentence}")
                        nok_sentences += 1

                now = time.time()

                if now - last_reported_time > REPORTING_INTERVAL_SECONDS:
                    logging.info(f"{ok_sentences} NMEA sentences logged, {nok_sentences} skipped.")
                    last_reported_time = now
                    ok_sentences = 0
                    nok_sentences = 0

                if now - last_rotation_time > ROTATION_INTERVAL_SECONDS:
                    logging.info(f"Rotation interval of {ROTATION_INTERVAL_SECONDS} seconds elapsed.")
                    break

if __name__ == "__main__":
    logging.basicConfig(
        level=logging.DEBUG,
        format='%(asctime)s %(levelname)s %(message)s',
        datefmt='%Y-%m-%dT%H:%M:%SZ'
    )

    logging.info(f"Starting NMEA logger")

    if not os.path.exists(OUTPUT_DIRECTORY):
        logging.info(f"Output directory [{OUTPUT_DIRECTORY}] does not exist, creating.")
        os.mkdir(OUTPUT_DIRECTORY)

    while True:
        try:
            with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
                logging.info(f"Connecting to {HOST}:{PORT}")
                s.connect((HOST, PORT))
                read_and_log(s)
        except ConnectionRefusedError:
            logging.error(f"Connection refused")
        except Exception as e:
            logging.exception(f"Exception occurred: {e}")
        finally:
            logging.info(f"Retrying after {RETRY_INTERVAL_SECONDS} second(s).")
            time.sleep(RETRY_INTERVAL_SECONDS)
