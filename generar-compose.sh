#!/bin/bash

OUTPUT_FILE=$1
NUM_CLIENTS=$2

echo "Nombre del archivo de salida: $OUTPUT_FILE"
echo "Cantidad de clientes: $NUM_CLIENTS"

python3 mi-generador.py $OUTPUT_FILE $NUM_CLIENTS