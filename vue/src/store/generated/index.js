// THIS FILE IS GENERATED AUTOMATICALLY. DO NOT MODIFY.
import RealiotechRealioNetworkRealiotechNetworkAsset from './realiotech/realio-network/realiotech.network.asset';
export default {
    RealiotechRealioNetworkRealiotechNetworkAsset: load(RealiotechRealioNetworkRealiotechNetworkAsset, 'realiotech.network.asset'),
};
function load(mod, fullns) {
    return function init(store) {
        if (store.hasModule([fullns])) {
            throw new Error('Duplicate module name detected: ' + fullns);
        }
        else {
            store.registerModule([fullns], mod);
            store.subscribe((mutation) => {
                if (mutation.type == 'common/env/INITIALIZE_WS_COMPLETE') {
                    store.dispatch(fullns + '/init', null, {
                        root: true
                    });
                }
            });
        }
    };
}
