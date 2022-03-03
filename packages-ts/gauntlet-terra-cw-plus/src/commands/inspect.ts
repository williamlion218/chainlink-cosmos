import { TerraCommand, TransactionResponse } from '@chainlink/gauntlet-terra'
import { Result } from '@chainlink/gauntlet-core'
import { logger } from '@chainlink/gauntlet-core/dist/utils'
import { Action, State, Vote } from '../lib/types'

export default class Inspect extends TerraCommand {
  static id = 'cw3_flex_multisig:inspect'

  constructor(flags, args: string[]) {
    super(flags, args)
  }

  makeRawTransaction = async () => {
    throw new Error('Query method does not have any tx')
  }

  fetchState = async (multisig: string, proposalId?: number): Promise<State> => {
    const query = this.provider.wasm.contractQuery.bind(this.provider.wasm)
    return fetchProposalState(query)(multisig, proposalId)
  }

  execute = async () => {
    const msig = this.args[0] || process.env.CW3_FLEX_MULTISIG
    const proposalId = Number(this.flags.proposal)
    const state = await this.fetchState(msig, proposalId)

    logger.info(makeInspectionMessage(state))
    return {} as Result<TransactionResponse>
  }
}

export const fetchProposalState = (query: (contractAddress: string, query: any) => Promise<any>) => async (
  multisig: string,
  proposalId?: number,
): Promise<State> => {
  const _queryMultisig = (params) => () => query(multisig, params)
  const multisigQueries = [
    _queryMultisig({
      list_voters: {},
    }),
    _queryMultisig({
      threshold: {},
    }),
  ]
  const proposalQueries = [
    _queryMultisig({
      proposal: {
        proposal_id: proposalId,
      },
    }),
    _queryMultisig({
      list_votes: {
        proposal_id: proposalId,
      },
    }),
  ]
  const queries = !!proposalId ? multisigQueries.concat(proposalQueries) : multisigQueries

  const [groupState, thresholdState, proposalState, votes] = await Promise.all(queries.map((q) => q()))

  const multisigState = {
    threshold: thresholdState.absolute_count.weight,
    owners: groupState.voters.map((m) => m.addr),
  }
  if (!proposalId) {
    return {
      multisig: multisigState,
      proposal: {
        nextAction: Action.CREATE,
        approvers: [],
      },
    }
  }
  const toNextAction = {
    passed: Action.EXECUTE,
    open: Action.APPROVE,
    pending: Action.APPROVE,
    rejected: Action.NONE,
    executed: Action.NONE,
  }
  return {
    multisig: multisigState,
    proposal: {
      id: proposalId,
      nextAction: toNextAction[proposalState.status],
      currentStatus: proposalState.status,
      data: proposalState.msgs,
      approvers: votes.votes.filter((v) => v.vote === Vote.YES).map((v) => v.voter),
      expiresAt: proposalState.expires.at_time ? new Date(proposalState.expires.at_time / 1e6) : null,
    },
  }
}

export const makeInspectionMessage = (state: State): string => {
  const newline = `\n`
  const indent = '  '.repeat(2)
  const ownersList = state.multisig.owners.map((o) => `\n${indent.repeat(2)} - ${o}`).join('')
  const multisigMessage = `Multisig State:
    - Threshold: ${state.multisig.threshold}
    - Total Owners: ${state.multisig.owners.length}
    - Owners List: ${ownersList}`

  let proposalMessage = `Proposal State:
    - Next Action: ${state.proposal.nextAction.toUpperCase()}`

  if (!state.proposal.id) return multisigMessage.concat(newline)

  const approversList = state.proposal.approvers.map((a) => `\n${indent.repeat(2)} - ${a}`).join('')
  proposalMessage = proposalMessage.concat(`
    - Proposal ID: ${state.proposal.id}
    - Total Approvers: ${state.proposal.approvers.length}
    - Approvers List: ${approversList}
    `)

  if (state.proposal.expiresAt) {
    const expiration = `- Approvals expire at ${state.proposal.expiresAt}`
    proposalMessage = proposalMessage.concat(expiration)
  }

  return multisigMessage.concat(newline).concat(proposalMessage).concat(newline)
}