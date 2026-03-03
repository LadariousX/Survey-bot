# code modified from Google Colab at
# https://github.com/Eric-Canas/QReader

# necessary dependency
# sudo apt-get install libzbar0

import sys
import os
from qreader import QReader
import cv2
import numpy as np

if len(sys.argv) < 2:
    print("Error: No image path provided.")
    sys.exit(1)

image_path = sys.argv[1]


try:
    #TODO: check valid file type
    with open(image_path, 'rb') as f:
        img_bytes = f.read()

    nparr = np.frombuffer(img_bytes, np.uint8)
    img = cv2.imdecode(nparr, cv2.IMREAD_COLOR)
    detector = QReader()
    # Detect and decode the QRs within the image
    decodedQRs, QRlocations = detector.detect_and_decode(image=img, return_detections=True)
    if len(decodedQRs) == 0:
        print("Error: No QR data could be decoded.")
        sys.exit(1)

    # Print the results
    for i, (decodedQR, QRlocation) in enumerate(zip(decodedQRs, QRlocations)):
        print(decodedQR)
        # print(f"QR {i+1} position: x: {QRlocation['cxcyn'][0]}, y: {QRlocation['cxcyn'][1]}")
        #print(f"Full detection info: {QRlocation}")
        sys.exit(0)
finally:
    # Always delete the image file, regardless of success or failure
    if os.path.exists(image_path):
        try:
            os.remove(image_path)
        except Exception as e:
            print(f"Warning: failed to delete image file {image_path}: {e}", file=sys.stderr)