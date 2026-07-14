# Notaion

| Abbreviation | Full name             | Explanation                                                                                |
| ------------ | --------------------- | ------------------------------------------------------------------------------------------ |
| Px           | Price                 |                                                                                            |
| Sz           | Size                  | In units of coin, i.e. base currency                                                       |
| Szi          | Signed size           | Positive for long, negative for short                                                      |
| Ntl          | Notional              | USD amount, Px \* Sz                                                                       |
| Side         | Side of trade or book | B = Bid = Buy, A = Ask = Short. Side is aggressing side for trades.                        |
| Asset        | Asset                 | An integer representing the asset being traded. See below for explanation                  |
| Tif          | Time in force         | GTC = good until canceled, ALO = add liquidity only (post only), IOC = immediate or cancel |

> Nonce adalah nomor urut yang tujuannya untuk mencegah pencurian aset (replay attack). Nomor ini harus urut secara kaku. Biasanya menggunakan timestamp (waktu saat ini dalam hitungan milidetik) sebagai angka Nonce, jadi otomatis selalu unik.

# API Wallet

Dikasih Wallet khusus untuk perwakilan khusus Bot (Agent Wallet) untuk melakukan signature transaksi/trading, Wallet ini berfungsi untuk melakukan aksi
Seperti (Open/Close) position. Untuk mengecek saldo/riwayat transaksi harus dari Main Wallet Address atau Sub-Account, bukan dari API Wallet address.
Kalo dipakai untuk cek saldo maka, datanya akan Kosong (Empty Result).
