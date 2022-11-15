import { BN, prompt } from '@chainlink/gauntlet-core/dist/utils'
import { AccAddress } from '@chainlink/gauntlet-terra'
import { CATEGORIES, ULUNA_DECIMALS } from '../../lib/constants'
import { TerraCommand, TransactionResponse, logger } from '@chainlink/gauntlet-terra'
import { Result } from '@chainlink/gauntlet-core'
import { withAddressBook } from '../../lib/middlewares'

import { MsgSendEncodeObject } from '@cosmjs/stargate'
import { MsgSend } from 'cosmjs-types/cosmos/bank/v1beta1/tx'

type CommandInput = {
  destination: string
  // Units in LUNA
  amount: string
}

export default class TransferLuna extends TerraCommand {
  static description = 'Transfer Luna'
  static examples = [`yarn gauntlet wallet:transfer --network=bombay-testnet`]

  static id = 'wallet:transfer'
  static category = CATEGORIES.WALLET

  input: CommandInput

  constructor(flags, args: string[]) {
    super(flags, args)
    this.use(withAddressBook)
  }

  buildCommand = async (flags, args): Promise<TerraCommand> => {
    this.input = this.makeInput(flags, args)
    return this
  }

  beforeExecute = async () => {
    logger.info(`Sending ${this.input.amount} uLUNA to ${logger.styleAddress(this.input.destination)}`)
  }

  makeInput = (flags, _) => {
    return {
      destination: flags.to,
      amount: new BN(flags.amount).mul(new BN(10).pow(new BN(ULUNA_DECIMALS))).toString(),
    } as CommandInput
  }

  makeRawTransaction = async (signer: AccAddress) => {
    if (!AccAddress.validate(this.input.destination)) throw new Error('Invalid destination address')

    const sendMsg: MsgSendEncodeObject = {
      typeUrl: "/cosmos.bank.v1beta1.MsgSend",
      value: MsgSend.fromPartial({
        fromAddress: signer,
        toAddress: this.input.destination,
        amount: [{ denom: 'ucosm', amount: this.input.amount }],
      }),
    };
    return [sendMsg]
  }

  execute = async () => {
    const message = await this.makeRawTransaction(this.signer.address)
    await this.beforeExecute()
    await prompt(`Continue?`)
    const tx = await this.signAndSend(message)
    const result = {
      responses: [
        {
          tx,
          contract: '',
        },
      ],
    } as Result<TransactionResponse>
    await this.afterExecute(result)
    return result
  }
}
