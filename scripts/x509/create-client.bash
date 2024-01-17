#!/bin/bash

set -ex

CA_DIR=.dev/tls
OUTPUT_DIR=${CA_DIR}/client
CURRENT_DIR=$(cd `dirname $0`; pwd)
config_file=${CURRENT_DIR}/openssl.cnf
clientname=ron

mkdir -p ${OUTPUT_DIR}
openssl genrsa -out ${OUTPUT_DIR}/${clientname}_key.pem 4096
openssl req -new \
  -key ${OUTPUT_DIR}/${clientname}_key.pem \
  -days 3650 \
  -out ${OUTPUT_DIR}/${clientname}_csr.pem \
  -subj /C=CN/ST=ShangHai/L=ShangHai/O=gRPC/CN=${clientname}/ \
  -config ${config_file} \
  -reqexts test_client
openssl x509 -req \
  -in ${OUTPUT_DIR}/${clientname}_csr.pem \
  -CAkey ${CA_DIR}/ca_key.pem \
  -CA ${CA_DIR}/ca_cert.pem \
  -days 3650 \
  -set_serial 1000 \
  -out ${OUTPUT_DIR}/${clientname}_cert.pem \
  -extfile ${config_file} \
  -extensions test_client \
  -sha256
openssl verify -verbose -CAfile ${CA_DIR}/ca_cert.pem  ${OUTPUT_DIR}/${clientname}_cert.pem

rm ${OUTPUT_DIR}/*_csr.pem
