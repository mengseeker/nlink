#!/bin/bash

set -ex

OUTPUT_DIR=.dev/tls
CURRENT_DIR=$(cd `dirname $0`; pwd)
config_file=${CURRENT_DIR}/openssl.cnf

# Create the server CA certs.
# openssl req -x509 \
#   -newkey rsa:4096 \
#   -nodes \
#   -days 3650 \
#   -keyout ${OUTPUT_DIR}/ca_key.pem \
#   -out ${OUTPUT_DIR}/ca_cert.pem \
#   -subj /C=CN/ST=ShangHai/L=ShangHai/O=gRPC/CN=test-server_ca/ \
#   -config ${config_file} \
#   -extensions test_ca \
#   -sha256

# Create the client CA certs.
# openssl req -x509 \
#   -newkey rsa:4096 \
#   -nodes \
#   -days 3650 \
#   -keyout ${OUTPUT_DIR}/client_ca_key.pem \
#   -out ${OUTPUT_DIR}/client_ca_cert.pem \
#   -subj /C=CN/ST=ShangHai/L=ShangHai/O=gRPC/CN=test-client_ca/ \
#   -config ${config_file} \
#   -extensions test_ca \
#   -sha256

# Generate a server cert.
openssl genrsa -out ${OUTPUT_DIR}/server_key.pem 4096
openssl req -new \
  -key ${OUTPUT_DIR}/server_key.pem \
  -days 3650 \
  -out ${OUTPUT_DIR}/server_csr.pem \
  -subj /C=CN/ST=ShangHai/L=ShangHai/O=gRPC/CN=example.com/ \
  -config ${config_file} \
  -reqexts test_server
openssl x509 -req \
  -in ${OUTPUT_DIR}/server_csr.pem \
  -CAkey ${OUTPUT_DIR}/ca_key.pem \
  -CA ${OUTPUT_DIR}/ca_cert.pem \
  -days 3650 \
  -set_serial 1000 \
  -out ${OUTPUT_DIR}/server_cert.pem \
  -extfile ${config_file} \
  -extensions test_server \
  -sha256
openssl verify -verbose -CAfile ${OUTPUT_DIR}/ca_cert.pem  ${OUTPUT_DIR}/server_cert.pem

rm ${OUTPUT_DIR}/*_csr.pem