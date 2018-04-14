package settings

//-----------------------------------------------------------------------------------------
//---------------Contains constants used througout the program-----------------------------
//-----------------------------------------------------------------------------------------

//---------------NET_FSM
const NO_FLOOR_FOUND = -1     //Used to show that no new order has been found
const ORDER_INACTIVE = 0      //Used to show that an order is not active in the timetable

const ORDER_TIMOUT_DELAY = 10 //[Seconds] The delay for when the system will dispatch another elevator to a floor
const PEERS_PORT = 14592      //Port used for peers system

//---------------ORDER_HANDLER
const DELEGATE_ORDER_PORT = 14002   //Port used for delegate order system
const NEW_ORDER_PORT = 14003        //Port used for new order system

const DELEGATE_ORDER_PACKET_LOSS_SIM_CHANCE = 70 //Chance of packetloss in delegate order system, for simulation
const NEW_ORDER_PACKET_LOSS_SIM_CHANCE = 30 //Chance of packetloss in new order system, for simulation

const SEND_ORDER_TIMEOUT_DELAY = 50 //[Milliseconds] Time until packet is regarded as lost and new packet sent
const SEND_ORDER_MAX_TIMEOUT_COUNT = 5 //Number of times send_order() will try to resend package

const TRANSMIT_ORDER_TO_MASTER_TIMEOUT_DELAY = 50 //[Milliseconds] Time until packet is regarded as lost and new packet sent
const TRANSMIT_ORDER_TO_MASTER_MAX_TIMOUT_COUNT = 5 //Number of times transmit_order_to_master() will try to resend package

//---------------SYNC
const SYNC_PORT = 16569 //Port used in sync system
const SYNC_TIMEOUT_DELAY = 50 //[Milliseconds] Time until packet is regarded as lost and new packet sent
const SYNC_MAX_TIMEOUT_COUNT = 5 //Number of times sync_state() will try to resend package
const SYNC_PACKET_LOSS_SIM_CHANCE = 25 //Chance of packetloss in sync system, for simulation

//--------------ELEV_FSM
const ELEV_CAB_TIMEOUT = 1000
