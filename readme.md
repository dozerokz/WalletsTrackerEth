# ETH Wallets Tracked (ERC-20)

This is a Go-based utility to monitor Ethereum wallets ERC-20 swaps and send notifications via Discord webhooks. It tracks ERC-20 transactions and internal transactions, providing detailed messages for buys/sells with token and ETH values.

The logic might not be perfect, but I hope this project can be useful to someone looking to track wallet transactions and trigger notifications.

(Only swaps from and to native cryptocurrencies supported. ETH in this case.)

![Webhook Examples](https://i.imgur.com/Qc9vVhY.png)


## Installation

1. Clone the repository:

``git clone https://github.com/dozerokz/walletsTrackerEth.git``

2. Navigate to the project directory:

``cd walletsTrackerEth``

3. Build the project:

``go build``

4. Run

``./main``

## Usage

1. Set up your Etherscan API key in the project. (In /walletsTracker/walletsTracker.go change line №16)

2. Define wallets you want to monitor. (Put in wallet.txt)

3. Configure Discord webhook URLs for notifications. (In /discordWebhook/webhook.go change line №8)

## License
This project is open-source. You can use, modify, and distribute it under the [MIT License](LICENSE).
