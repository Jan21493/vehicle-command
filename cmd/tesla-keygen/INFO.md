
# Decoding output from tesla-control list-keys and compare with stored public key

output from: tesla-control list-keys | jq
{
  "rssi": -72,
  "keylist": [
...
    {
      "publicKey": "043857028850ea9307ed3657f4375a60bf06b86b9c6a7786562b6cf5ae4afd9842a8d1d02f86edea45ecf4d368231d35bb47447c83d84370b2903ca0ab1665abbe",
      "role": "ROLE_OWNER",
      "formFactor": "KEY_FORM_FACTOR_CLOUD_KEY"
    },
...
  ]
}

public key in PEM format:
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEOFcCiFDqkwftNlf0N1pgvwa4a5xq
d4ZWK2z1rkr9mEKo0dAvhu3qRez002gjHTW7R0R8g9hDcLKQPKCrFmWrvg==
-----END PUBLIC KEY-----

converted to hex via https://cryptii.com/pipes/base64-to-hex (format hex, no grouping)
input: MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEOFcCiFDqkwftNlf0N1pgvwa4a5xqd4ZWK2z1rkr9mEKo0dAvhu3qRez002gjHTW7R0R8g9hDcLKQPKCrFmWrvg==
output: 3059301306072a8648ce3d020106082a8648ce3d030107034200043857028850ea9307ed3657f4375a60bf06b86b9c6a7786562b6cf5ae4afd9842a8d1d02f86edea45ecf4d368231d35bb47447c83d84370b2903ca0ab1665abbe

decoded via http://ldh.org/asn1.html (hex to ASN.1)

U.P.SEQUENCE {
   U.P.SEQUENCE {
      U.P.OBJECTIDENTIFIER 1.2.840.10045.2.1 (ecPublicKey)
      U.P.OBJECTIDENTIFIER 1.2.840.10045.3.1.7 (P-256)
   }
   U.P.BITSTRING         # 00043857028850ea9307ed3657f4375a60bf06b86b9c6a7786562b6cf5ae4afd9842a8d1d02f86edea45ecf4d368231d35bb47447c83d84370b2903ca0ab1665abbe
043857028850ea9307ed3657f4375a60bf06b86b9c6a7786562b6cf5ae4afd9842a8d1d02f86edea45ecf4d368231d35bb47447c83d84370b2903ca0ab1665abbe : 0 unused bit(s); 
}

key findings:
- publicKey output from tesla-control list-keys is the same as the last 65 bytes from the converted and decoded ASN.1 message

###########
base64Key: MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEOFcCiFDqkwftNlf0N1pgvwa4a5xqd4ZWK2z1rkr9mEKo0dAvhu3qRez002gjHTW7R0R8g9hDcLKQPKCrFmWrvg==

3059301306072a8648ce3d020106082a8648ce3d030107034200043857028850ea9307ed3657f4375a60bf06b86b9c6a7786562b6cf5ae4afd9842a8d1d02f86edea45ecf4d368231d35bb47447c83d84370b2903ca0ab1665abbe

Jans Tesla 