## Notaion

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

## API Wallet

Dikasih Wallet khusus untuk perwakilan khusus Bot (Agent Wallet) untuk melakukan signature transaksi/trading, Wallet ini berfungsi untuk melakukan aksi
Seperti (Open/Close) position. Untuk mengecek saldo/riwayat transaksi harus dari Main Wallet Address atau Sub-Account, bukan dari API Wallet address.
Kalo dipakai untuk cek saldo maka, datanya akan Kosong (Empty Result).

## Info Endpoint

Untuk fetch data seperti:

- Melihat Order Book
- Mengecek Status Order
- Melihat Riwayat Transaksi
- Mengecek Harga Koin Saat ini

> Note: Hanya butuh POST request aja menggunakan metode HTTP `POST` ke endpoint: https://api.hyperliquid.xyz/info

| Tipe Request (`type`) | Fungsi Utama                                                                             | Contoh Payload (JSON)                                    | Aturan Penting                                                                                                                                             |
| --------------------- | ---------------------------------------------------------------------------------------- | -------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `allMids`             | Cek harga tengah (_mid price_) seluruh koin secara instan.                               | `{"type": "allMids"}`                                    | Mengembalikan data _key-value_ harga koin terbaru. Sangat ringan dan cepat dibaca bot.                                                                     |
| `l2Book`              | Mengintip antrean order (_order book L2_) maksimal 20 baris per sisi.                    | `{"type": "l2Book", "coin": "BTC", "nSigFigs": null}`    | Masukkan nama koin (_perp_) atau indeks `@` (_spot_). Isi `nSigFigs` dengan nilai `2–5` untuk membulatkan desimal agar bot lebih cepat memproses hitungan. |
| `orderStatus`         | Melacak status order: sukses, masih antre, atau alasan spesifik kenapa dibatalkan/gagal. | `{"type": "orderStatus", "user": "0x...", "oid": 12345}` | Wajib menggunakan alamat **Main Wallet**. Mengembalikan alasan detail seperti `reduceOnlyRejected` atau `selfTradeCanceled`.                               |
| `userFills`           | Melihat riwayat transaksi trading yang sudah berhasil dieksekusi (_done_).               | `{"type": "userFills", "user": "0x..."}`                 | Maksimal mengambil **2.000 data**. Jika membutuhkan filter waktu, gunakan `userFillsByTime` dengan parameter tambahan `startTime`.                         |
