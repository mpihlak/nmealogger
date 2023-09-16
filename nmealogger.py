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
        # TODO: Since my Raspberry doesn't have the correct time, switch to
        # sequence number based naming instead. This will also make the file names
        # easier to handle with tools such as grep, cut and awk.
        file_ts = datetime.datetime.now().strftime('%Y-%m-%dT%H:%M:%S')
        filename = os.path.join(OUTPUT_DIRECTORY, "nmea-{file_ts}.log".format(file_ts=file_ts))
        logging.info("Writing to " + filename)

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
                    logging.warning("Data does not end in newline: " + strdata)
                else:
                    strdata = strdata.rstrip()
                sentences = strdata.split('\n')

                ts = datetime.datetime.now(datetime.timezone.utc).isoformat()
                for sentence in sentences:
                    sentence = sentence.strip("\r")
                    if valid_nmea_checksum(sentence):
                        f.write((ts + "\t" + sentence + "\n").encode('ascii'))
                        f.flush()
                        ok_sentences += 1
                    else:
                        logging.warning("Skipping sentence with invalid checksum: " + sentence)
                        nok_sentences += 1

                now = time.time()

                if now - last_reported_time > REPORTING_INTERVAL_SECONDS:
                    logging.info("{ok_sentences} NMEA sentences logged, {nok_sentences} skipped.".format(ok_sentences=ok_sentences, nok_sentences=nok_sentences))
                    last_reported_time = now
                    ok_sentences = 0
                    nok_sentences = 0

                if now - last_rotation_time > ROTATION_INTERVAL_SECONDS:
                    logging.info("File rotation interval elapsed.")
                    break

if __name__ == "__main__":
    logging.basicConfig(
        level=logging.DEBUG,
        format='%(asctime)s %(levelname)s %(message)s',
        datefmt='%Y-%m-%dT%H:%M:%SZ'
    )

    logging.info("Starting NMEA logger")

    if not os.path.exists(OUTPUT_DIRECTORY):
        logging.info("Output directory does not exist, creating " + OUTPUT_DIRECTORY)
        os.mkdir(OUTPUT_DIRECTORY)

    while True:
        try:
            with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
                logging.info("Connecting to {HOST}:{PORT}".format(HOST=HOST,PORT=PORT))
                s.connect((HOST, PORT))
                read_and_log(s)
        except ConnectionRefusedError:
            logging.error("Connection refused")
        except Exception as e:
            logging.exception("Exception occurred: " + str(e))
        finally:
            logging.info("Retrying ...")
            time.sleep(RETRY_INTERVAL_SECONDS)
