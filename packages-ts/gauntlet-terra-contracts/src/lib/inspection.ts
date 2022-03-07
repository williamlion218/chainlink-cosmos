import { Proto } from '@chainlink/gauntlet-core/dist/crypto'
import { BN } from '@chainlink/gauntlet-core/dist/utils'
import { AccAddress, LCDClient } from '@terra-money/terra.js'
import { providerUtils } from '@chainlink/gauntlet-terra'
import { logger } from '@chainlink/gauntlet-core/dist/utils'
import { deepCopy } from './utils'
import Long from 'long'

// TODO: find the right place for this function
export const getLatestOCRConfigEvent = async (provider: LCDClient, contract: AccAddress) => {
  // The contract only stores the block where the config was accepted. The tx log contains the config
  const latestConfigDetails: any = await provider.wasm.contractQuery(contract, 'latest_config_details' as any)
  const setConfigTx = providerUtils.filterTxsByEvent(
    await providerUtils.getBlockTxs(provider, latestConfigDetails.block_number),
    'wasm-set_config',
  )

  return setConfigTx?.logs?.[0].eventsByType['wasm-set_config']
}

enum DIFF_PROPERTY_COLOR {
  ADDED = 'green',
  REMOVED = 'red',
  NO_CHANGE = 'reset',
}

type DIFF_OPTIONS = {
  initialIndent?: string
  propertyName?: string
}

// TODO: find a better place for this function (likely gauntlet-core to expose it for other project)
// https://github.com/smartcontractkit/chainlink-terra/issues/181
export function printDiff(existing: Object, incoming: Object, options?: DIFF_OPTIONS) {
  const { initialIndent = '', propertyName = 'Object' } = options || {}
  logger.log(initialIndent, propertyName, '{')
  const indent = initialIndent + '  '

  for (const prop of Object.keys(incoming)) {
    const existingProperty = existing?.[prop]
    const incomingProperty = incoming[prop]

    if (Array.isArray(incomingProperty)) {
      logger.log(indent, prop, ': [')
      const itemsIndent = indent + '  '

      for (const item of incomingProperty) {
        const itemStr = Buffer.isBuffer(item) ? item.toString('hex') : item
        if (existingProperty?.includes(item)) {
          logger.log(itemsIndent, logger.style(itemStr, DIFF_PROPERTY_COLOR.NO_CHANGE))
        } else {
          logger.log(itemsIndent, logger.style(itemStr, DIFF_PROPERTY_COLOR.ADDED))
        }
      }

      for (const item of existingProperty || []) {
        const itemStr = Buffer.isBuffer(item) ? item.toString('hex') : item
        if (!incomingProperty.includes(item)) {
          logger.log(itemsIndent, logger.style(itemStr, DIFF_PROPERTY_COLOR.REMOVED))
        }
      }
      logger.log(indent, `]`)
      continue
    }

    if (Buffer.isBuffer(incomingProperty)) {
      if (Buffer.compare(incomingProperty, existingProperty || Buffer.from('')) === 0) {
        logger.log(indent, `${prop}:`, logger.style(incomingProperty.toString('hex'), DIFF_PROPERTY_COLOR.NO_CHANGE))
      } else {
        logger.log(indent, `${prop}:`, logger.style(existingProperty?.toString('hex'), DIFF_PROPERTY_COLOR.REMOVED))
        logger.log(indent, `${prop}:`, logger.style(incomingProperty.toString('hex'), DIFF_PROPERTY_COLOR.ADDED))
      }
      continue
    }

    if (typeof incomingProperty === 'object') {
      printDiff(existingProperty, incomingProperty, {
        initialIndent: indent,
        propertyName: `${prop}:`,
      })
      continue
    }

    // plain property
    if (existingProperty == incomingProperty) {
      logger.log(indent, `${prop}:`, logger.style(incomingProperty, DIFF_PROPERTY_COLOR.NO_CHANGE))
    } else {
      logger.log(indent, `${prop}:`, logger.style(existingProperty, DIFF_PROPERTY_COLOR.REMOVED))
      logger.log(indent, `${prop}:`, logger.style(incomingProperty, DIFF_PROPERTY_COLOR.ADDED))
    }
  }

  logger.log(initialIndent, '}')
}

export const longsInObjToNumbers = (obj) => {
  const copy = deepCopy(obj)
  for (const [key, value] of Object.entries(obj)) {
    if (Array.isArray(value) || Buffer.isBuffer(value) || value instanceof Date) {
      // skip non-convertable arrays and buffers
      continue
    }

    if (Long.isLong(value)) {
      // transform long struct into readable and comparable number
      copy[key] = toComparableLongNumber(value)
      continue
    }

    if (typeof value === 'object') {
      // for all nested objects repeat recursively
      copy[key] = longsInObjToNumbers(value)
    }
  }
  return copy
}

export const toComparableLongNumber = (v: Long) => new BN(Proto.Protobuf.longToString(v)).toString()

export const toComparableNumber = (v: string | number | Long) => {
  // Proto encoding will ignore falsy values
  if (!v) return '0'
  if (typeof v === 'string' || typeof v === 'number') return new BN(v).toString()
  return toComparableLongNumber(v)
}