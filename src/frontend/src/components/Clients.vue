<template>
    <v-container>
        <v-card>
            <v-card-title>
                Clients
                <v-switch
                        class="ml-3"
                        dark
                        :label="listView ? 'Switch to card view' : 'Switch to list view'"
                        v-model="listView"
                />
                <v-spacer></v-spacer>
                <v-text-field
                        v-if="listView"
                        v-model="search"
                        append-icon="mdi-magnify"
                        label="Search"
                        single-line
                        hide-details
                ></v-text-field>
                <v-spacer></v-spacer>
                <v-btn
                        color="success"
                        @click="startCreate"
                >
                    Add new client
                    <v-icon right dark>mdi-account-multiple-plus-outline</v-icon>
                </v-btn>
            </v-card-title>
            <v-data-table
                    v-if="listView"
                    :headers="headers"
                    :items="clients"
                    :search="search"
            >
                <template v-slot:item.address="{ item }">
                    <v-chip
                            v-for="(ip, i) in item.address"
                            :key="i"
                            color="indigo"
                            text-color="white"
                    >
                        <v-icon left>mdi-ip-network</v-icon>
                        {{ ip }}
                    </v-chip>
                </template>
                <template v-slot:item.tags="{ item }">
                    <v-chip
                            v-for="(tag, i) in item.tags"
                            :key="i"
                            color="blue-grey"
                            text-color="white"
                    >
                        <v-icon left>mdi-tag</v-icon>
                        {{ tag }}
                    </v-chip>
                </template>
                <template v-slot:item.created="{ item }">
                    <v-row>
                        <p>At {{ item.created | formatDate }} by {{ item.createdBy }}</p>
                    </v-row>
                </template>
                <template v-slot:item.updated="{ item }">
                    <v-row>
                        <p>At {{ item.updated | formatDate }} by {{ item.updatedBy }}</p>
                    </v-row>
                </template>
                <template v-slot:item.action="{ item }">
                    <v-row>
                        <v-icon
                                class="pr-1 pl-1"
                                @click.stop="startUpdate(item)"
                        >
                            mdi-square-edit-outline
                        </v-icon>
                        <v-icon
                                class="pr-1 pl-1"
                                @click.stop="forceFileDownload(item)"
                        >
                            mdi-cloud-download-outline
                        </v-icon>
                        <v-icon
                                class="pr-1 pl-1"
                                @click.stop="email(item)"
                        >
                            mdi-email-send-outline
                        </v-icon>
                        <v-icon
                                class="pr-1 pl-1"
                                @click="remove(item)"
                        >
                            mdi-trash-can-outline
                        </v-icon>
                        <v-switch
                                dark
                                class="pr-1 pl-1"
                                color="success"
                                v-model="item.enable"
                                v-on:change="update(item)"
                        />
                    </v-row>
                </template>

            </v-data-table>
            <v-card-text v-else>
                <v-row>
                    <v-col
                            v-for="(client, i) in clients"
                            :key="i"
                            sm12 lg6
                    >
                        <v-card
                                :color="client.enable ? '#1F7087' : 'warning'"
                                class="mx-auto"
                                raised
                                shaped
                        >
                            <v-list-item>
                                <v-list-item-content>
                                    <v-list-item-title class="headline">{{ client.name }}</v-list-item-title>
                                    <v-list-item-subtitle>{{ client.email }}</v-list-item-subtitle>
                                    <v-list-item-subtitle>Created: {{ client.created | formatDate }} by {{ client.createdBy }}</v-list-item-subtitle>
                                    <v-list-item-subtitle>Updated: {{ client.updated | formatDate }} by {{ client.updatedBy }}</v-list-item-subtitle>
                                </v-list-item-content>

                                <v-list-item-avatar
                                        tile
                                        size="150"
                                >
                                    <v-img :src="'data:image/png;base64, ' + getClientQrcode(client.id)"/>
                                </v-list-item-avatar>
                            </v-list-item>

                            <v-card-text class="text--primary">
                                <v-chip
                                        v-for="(ip, i) in client.address"
                                        :key="i"
                                        color="indigo"
                                        text-color="white"
                                >
                                    <v-icon left>mdi-ip-network</v-icon>
                                    {{ ip }}
                                </v-chip>
                            </v-card-text>
                            <v-card-text class="text--primary">
                                <v-chip
                                        v-for="(tag, i) in client.tags"
                                        :key="i"
                                        color="blue-grey"
                                        text-color="white"
                                >
                                    <v-icon left>mdi-tag</v-icon>
                                    {{ tag }}
                                </v-chip>
                            </v-card-text>
                            <v-card-actions>
                                <v-tooltip bottom>
                                    <template v-slot:activator="{ on }">
                                        <v-btn
                                                text
                                                v-on:click="forceFileDownload(client)"
                                                v-on="on"
                                        >
                                            <span class="d-none d-lg-flex">Download</span>
                                            <v-icon right dark>mdi-cloud-download-outline</v-icon>
                                        </v-btn>
                                    </template>
                                    <span>Download</span>
                                </v-tooltip>

                                <v-tooltip bottom>
                                    <template v-slot:activator="{ on }">
                                        <v-btn
                                                text
                                                @click.stop="startUpdate(client)"
                                                v-on="on"
                                        >
                                            <span class="d-none d-lg-flex">Edit</span>
                                            <v-icon right dark>mdi-square-edit-outline</v-icon>
                                        </v-btn>
                                    </template>
                                    <span>Edit</span>
                                </v-tooltip>

                                <v-tooltip bottom>
                                    <template v-slot:activator="{ on }">
                                        <v-btn
                                                text
                                                @click="remove(client)"
                                                v-on="on"
                                        >
                                            <span class="d-none d-lg-flex">Delete</span>
                                            <v-icon right dark>mdi-trash-can-outline</v-icon>
                                        </v-btn>
                                    </template>
                                    <span>Delete</span>
                                </v-tooltip>

                                <v-tooltip bottom>
                                    <template v-slot:activator="{ on }">
                                        <v-btn
                                                text
                                                @click="email(client)"
                                                v-on="on"
                                        >
                                            <span class="d-none d-lg-flex">Send Email</span>
                                            <v-icon right dark>mdi-email-send-outline</v-icon>
                                        </v-btn>
                                    </template>
                                    <span>Send Email</span>
                                </v-tooltip>
                                <v-spacer/>
                                <v-tooltip right>
                                    <template v-slot:activator="{ on }">
                                        <v-switch
                                                dark
                                                v-on="on"
                                                color="success"
                                                v-model="client.enable"
                                                v-on:change="update(client)"
                                        />
                                    </template>
                                    <span> {{client.enable ? 'Disable' : 'Enable'}} this client</span>
                                </v-tooltip>

                            </v-card-actions>
                        </v-card>
                    </v-col>
                </v-row>
            </v-card-text>
        </v-card>
        <v-dialog
                v-if="client"
                v-model="dialogCreate"
                max-width="550"
        >
            <v-card>
                <v-card-title class="headline">Add new client</v-card-title>
                <v-card-text>
                    <v-row>
                        <v-col
                                cols="12"
                        >
                            <v-form
                                    ref="form"
                                    v-model="valid"
                            >
                                <v-text-field
                                        v-model="client.name"
                                        label="Client friendly name"
                                        :rules="[ v => !!v || 'Client name is required', ]"
                                        required
                                />
                                <v-text-field
                                        v-model="client.email"
                                        label="Client email"
                                        :rules="[ v => (/.+@.+\..+/.test(v) || v === '') || 'E-mail must be valid',]"
                                />
                                <v-select
                                        v-model="client.address"
                                        :items="server.address"
                                        label="Client IP will be chosen from these networks"
                                        :rules="[ v => !!v || 'Network is required', ]"
                                        multiple
                                        chips
                                        persistent-hint
                                        required
                                />
                                <v-combobox
                                        v-model="client.allowedIPs"
                                        chips
                                        hint="Write IPv4 or IPv6 CIDR and hit enter"
                                        label="Allowed IPs"
                                        multiple
                                        dark
                                >
                                    <template v-slot:selection="{ attrs, item, select, selected }">
                                        <v-chip
                                                v-bind="attrs"
                                                :input-value="selected"
                                                close
                                                @click="select"
                                                @click:close="client.allowedIPs.splice(client.allowedIPs.indexOf(item), 1)"
                                        >
                                            <strong>{{ item }}</strong>&nbsp;
                                        </v-chip>
                                    </template>
                                </v-combobox>
                                <v-combobox
                                        v-model="client.tags"
                                        chips
                                        hint="Write tag name and hit enter"
                                        label="Tags"
                                        multiple
                                        dark
                                >
                                    <template v-slot:selection="{ attrs, item, select, selected }">
                                        <v-chip
                                                v-bind="attrs"
                                                :input-value="selected"
                                                close
                                                @click="select"
                                                @click:close="client.tags.splice(client.tags.indexOf(item), 1)"
                                        >
                                            <strong>{{ item }}</strong>&nbsp;
                                        </v-chip>
                                    </template>
                                </v-combobox>
                                <v-switch
                                        v-model="client.enable"
                                        color="red"
                                        inset
                                        :label="client.enable ? 'Enable client after creation': 'Disable client after creation'"
                                />
                                <v-switch
                                        v-model="client.ignorePersistentKeepalive"
                                        color="red"
                                        inset
                                        :label="'Ignore global persistent keepalive: ' + (client.ignorePersistentKeepalive ? 'Yes': 'NO')"
                                        @change="handlePersistentKeepaliveChange"
                                />
                                <template v-if="client.ignorePersistentKeepalive">
                                    <v-switch
                                            v-model="client.keepaliveDisabled"
                                            color="red"
                                            inset
                                            :label="'Persist Keepalive Disabled: ' + (client.keepaliveDisabled ? 'Yes': 'NO')"
                                            @change="handleKeepaliveDisabledChange"
                                    />

                                    <v-text-field
                                            v-if="!client.keepaliveDisabled"
                                            v-model.number="client.keepaliveInterval"
                                            label="Persist Keepalive Interval"
                                            type="number"
                                            min="1"
                                            :rules="[
                                                v => (client.ignorePersistentKeepalive && !client.keepaliveDisabled ? !!v : true) || 'Keepalive interval is required when keepalive is enabled',
                                                v => (client.ignorePersistentKeepalive && !client.keepaliveDisabled ? v > 0 : true) || 'Interval must be greater than 0'
                                            ]"
                                            hint="Interval in seconds"
                                            persistent-hint
                                    />
                                </template>
                                <v-switch
                                        v-model="client.useRemoteDNS"
                                        color="green"
                                        inset
                                        :label="'Use server DNS: ' + (client.useRemoteDNS ? 'Yes': 'NO')"
                                />                                

                                <v-switch
                                        v-model="client.site2site"
                                        color="red"
                                        inset
                                        :label="'Site-to-Site Client: ' + (client.site2site ? 'Yes': 'NO')"
                                        @change="handleSite2SiteChange"
                                />
                                <template v-if="client.site2site">
                                          <v-combobox
                                                v-model="client.lanIPs"
                                                chips
                                                hint="Client side LAN subnets to route over VPN"
                                                label="Client LAN IP's"
                                                multiple
                                                dark
                                                :rules="[v => !!v || 'LAN IPs are required when Site-to-Site is enabled']"
                                            >
                                                <template v-slot:selection="{ attrs, item, select, selected }">
                                                    <v-chip
                                                        v-bind="attrs"
                                                        :input-value="selected"
                                                        close
                                                        @click="select"
                                                        @click:close="client.lanIPs.splice(client.lanIPs.indexOf(item), 1)"
                                                    >
                                                        <strong>{{ item }}</strong>&nbsp;
                                                    </v-chip>
                                                </template>
                                            </v-combobox>

                                            <v-select
                                                v-model="client.table"
                                                :items="tableOptions"
                                                label="WireGuard routing table"
                                                hint="Optionally define routing table to use for Wireguard, or leave as AUTO (Default) or OFF to disable automatic creation of routes"
                                                persistent-hint
                                                :return-object="false"
                                                :menu-props="{ closeOnContentClick: false }"
                                            >
                                                <template v-slot:prepend-item>
                                                    <v-list-item>
                                                        <v-text-field
                                                            v-model="customTableValue"
                                                            label="Custom table value (integer)"
                                                            type="number"
                                                            @keydown.enter="setCustomTableValue"
                                                        />
                                                        <v-btn
                                                            small
                                                            color="primary"
                                                            class="ml-2"
                                                            @click="setCustomTableValue"
                                                        >
                                                            Apply
                                                        </v-btn>
                                                    </v-list-item>
                                                    <v-divider></v-divider>
                                                </template>
                                            </v-select>

                                            <v-switch
                                                v-model="client.Site2SiteEndpointOptionsEnabled"
                                                color="blue"
                                                inset
                                                :label="'Enable Endpoint Options: ' + (client.Site2SiteEndpointOptionsEnabled ? 'Yes' : 'No')"
                                                @change="handleEnableEndpointOptionsChange"
                                            />

                                            <template v-if="client.Site2SiteEndpointOptionsEnabled">
                                                <v-text-field
                                                    v-model="client.Site2SiteEndpoint"
                                                    label="Client endpoint address"
                                                    hint="IP address or URL of client endpoint if open to internet"
                                                    persistent-hint
                                                />
                                                <v-text-field
                                                    v-model.number="client.Site2SiteEndpointListenPort"
                                                    label="Client endpoint listen port"
                                                    type="number"
                                                    hint="Listen port to use if client open to internet"
                                                    persistent-hint
                                                    :rules="[v => (!client.Site2SiteEndpoint || !!v) || 'Listen port is required if endpoint is set']"
                                                />
                                                <v-text-field
                                                    v-model.number="client.Site2SiteEndpointPort"
                                                    label="Client endpoint public port"
                                                    type="number"
                                                    hint="Public port for client endpoint if DNAT is being used"
                                                    persistent-hint
                                                    :rules="[v => (!client.Site2SiteEndpoint || !v || v > 0) || 'Public port must be a positive integer']"
                                                    @input="validatePortInput($event, 'Site2SiteEndpointPort')"
                                                />
                                            </template>
                                </template>  

                            </v-form>
                        </v-col>
                    </v-row>
                </v-card-text>
                <v-card-actions>
                    <v-spacer/>
                    <v-btn
                            :disabled="!valid"
                            color="success"
                            @click="create(client)"
                    >
                        Submit
                        <v-icon right dark>mdi-check-outline</v-icon>
                    </v-btn>
                    <v-btn
                            color="primary"
                            @click="dialogCreate = false"
                    >
                        Cancel
                        <v-icon right dark>mdi-close-circle-outline</v-icon>
                    </v-btn>
                </v-card-actions>
            </v-card>
        </v-dialog>
        <v-dialog
                v-if="client"
                v-model="dialogUpdate"
                max-width="550"
        >
            <v-card>
                <v-card-title class="headline">Edit client</v-card-title>
                <v-card-text>
                    <v-row>
                        <v-col
                                cols="12"
                        >
                            <v-form
                                    ref="form"
                                    v-model="valid"
                            >
                                <v-text-field
                                        v-model="client.name"
                                        label="Friendly name"
                                        :rules="[ v => !!v || 'Client name is required',]"
                                        required
                                />
                                <v-text-field
                                        v-model="client.email"
                                        label="Email"
                                        :rules="[ v => (/.+@.+\..+/.test(v) || v === '') || 'E-mail must be valid',]"
                                        required
                                />
                                <v-combobox
                                        v-model="client.address"
                                        chips
                                        hint="Write IPv4 or IPv6 CIDR and hit enter"
                                        label="Addresses"
                                        multiple
                                        dark
                                >
                                    <template v-slot:selection="{ attrs, item, select, selected }">
                                        <v-chip
                                                v-bind="attrs"
                                                :input-value="selected"
                                                close
                                                @click="select"
                                                @click:close="client.address.splice(client.address.indexOf(item), 1)"
                                        >
                                            <strong>{{ item }}</strong>&nbsp;
                                        </v-chip>
                                    </template>
                                </v-combobox>
                                <v-combobox
                                        v-model="client.allowedIPs"
                                        chips
                                        hint="Write IPv4 or IPv6 CIDR and hit enter"
                                        label="Allowed IPs"
                                        multiple
                                        dark
                                >
                                    <template v-slot:selection="{ attrs, item, select, selected }">
                                        <v-chip
                                                v-bind="attrs"
                                                :input-value="selected"
                                                close
                                                @click="select"
                                                @click:close="client.allowedIPs.splice(client.allowedIPs.indexOf(item), 1)"
                                        >
                                            <strong>{{ item }}</strong>&nbsp;
                                        </v-chip>
                                    </template>
                                </v-combobox>
                                <v-combobox
                                        v-model="client.tags"
                                        chips
                                        hint="Write tag name and hit enter"
                                        label="Tags"
                                        multiple
                                        dark
                                >
                                    <template v-slot:selection="{ attrs, item, select, selected }">
                                        <v-chip
                                                v-bind="attrs"
                                                :input-value="selected"
                                                close
                                                @click="select"
                                                @click:close="client.tags.splice(client.tags.indexOf(item), 1)"
                                        >
                                            <strong>{{ item }}</strong>&nbsp;
                                        </v-chip>
                                    </template>
                                </v-combobox>
                                <v-switch
                                        v-model="client.ignorePersistentKeepalive"
                                        color="red"
                                        inset
                                        :label="'Ignore global persistent keepalive: ' + (client.ignorePersistentKeepalive ? 'Yes': 'NO')"
                                        @change="handlePersistentKeepaliveChange"
                                />
                                <template v-if="client.ignorePersistentKeepalive">
                                    <v-switch
                                            v-model="client.keepaliveDisabled"
                                            color="red"
                                            inset
                                            :label="'Persist Keepalive Disabled: ' + (client.keepaliveDisabled ? 'Yes': 'NO')"
                                            @change="handleKeepaliveDisabledChange"
                                    />

                                    <v-text-field
                                            v-if="!client.keepaliveDisabled"
                                            v-model.number="client.keepaliveInterval"
                                            label="Persist Keepalive Interval"
                                            type="number"
                                            min="1"
                                            :rules="[
                                                v => (client.ignorePersistentKeepalive && !client.keepaliveDisabled ? !!v : true) || 'Keepalive interval is required when keepalive is enabled',
                                                v => (client.ignorePersistentKeepalive && !client.keepaliveDisabled ? v > 0 : true) || 'Interval must be greater than 0'
                                            ]"
                                            hint="Interval in seconds"
                                            persistent-hint
                                    />
                                </template>
                                <v-switch
                                        v-model="client.useRemoteDNS"
                                        color="green"
                                        inset
                                        :label="'Use server DNS: ' + (client.useRemoteDNS ? 'Yes': 'NO')"
                                />
                                <v-switch
                                        v-model="client.site2site"
                                        color="red"
                                        inset
                                        :label="'Site-to-Site Client: ' + (client.site2site ? 'Yes': 'NO')"
                                        @change="handleSite2SiteChange"
                                />
                                <template v-if="client.site2site">
                                          <v-combobox
                                                v-model="client.lanIPs"
                                                chips
                                                hint="Client side LAN subnets to route over VPN"
                                                label="Client LAN IP's"
                                                multiple
                                                dark
                                                :rules="[v => !!v || 'LAN IPs are required when Site-to-Site is enabled']"
                                            >
                                                <template v-slot:selection="{ attrs, item, select, selected }">
                                                    <v-chip
                                                        v-bind="attrs"
                                                        :input-value="selected"
                                                        close
                                                        @click="select"
                                                        @click:close="client.lanIPs.splice(client.lanIPs.indexOf(item), 1)"
                                                    >
                                                        <strong>{{ item }}</strong>&nbsp;
                                                    </v-chip>
                                                </template>
                                            </v-combobox>

                                            <v-select
                                                v-model="client.table"
                                                :items="tableOptions"
                                                label="WireGuard routing table"
                                                hint="Optionally define routing table to use for Wireguard, or leave as AUTO (Default) or OFF to disable automatic creation of routes"
                                                persistent-hint
                                                :return-object="false"
                                                :menu-props="{ closeOnContentClick: false }"
                                            >
                                                <template v-slot:prepend-item>
                                                    <v-list-item>
                                                        <v-text-field
                                                            v-model="customTableValue"
                                                            label="Custom table value (integer)"
                                                            type="number"
                                                            @keydown.enter="setCustomTableValue"
                                                        />
                                                        <v-btn
                                                            small
                                                            color="primary"
                                                            class="ml-2"
                                                            @click="setCustomTableValue"
                                                        >
                                                            Apply
                                                        </v-btn>
                                                    </v-list-item>
                                                    <v-divider></v-divider>
                                                </template>
                                            </v-select>

                                            <v-switch
                                                v-model="client.Site2SiteEndpointOptionsEnabled"
                                                color="blue"
                                                inset
                                                :label="'Enable Endpoint Options: ' + (client.Site2SiteEndpointOptionsEnabled ? 'Yes' : 'No')"
                                                @change="handleEnableEndpointOptionsChange"
                                            />

                                            <template v-if="client.Site2SiteEndpointOptionsEnabled">
                                                <v-text-field
                                                    v-model="client.Site2SiteEndpoint"
                                                    label="Client endpoint address"
                                                    hint="IP address or URL of client endpoint if open to internet"
                                                    persistent-hint
                                                />
                                                <v-text-field
                                                    v-model.number="client.Site2SiteEndpointListenPort"
                                                    label="Client endpoint listen port"
                                                    type="number"
                                                    hint="Listen port to use if client open to internet"
                                                    persistent-hint
                                                    :rules="[v => (!client.Site2SiteEndpoint || !!v) || 'Listen port is required if endpoint is set']"
                                                />
                                                <v-text-field
                                                    v-model.number="client.Site2SiteEndpointPort"
                                                    label="Client endpoint public port"
                                                    type="number"
                                                    hint="Public port for client endpoint if DNAT is being used"
                                                    persistent-hint
                                                    :rules="[v => (!client.Site2SiteEndpoint || !v || v > 0) || 'Public port must be a positive integer']"
                                                    @input="validatePortInput($event, 'Site2SiteEndpointPort')"
                                                />
                                            </template>
                                </template>  

                            </v-form>
                        </v-col>
                    </v-row>
                </v-card-text>
                <v-card-actions>
                    <v-spacer/>
                    <v-btn
                            :disabled="!valid"
                            color="success"
                            @click="update(client)"
                    >
                        Submit
                        <v-icon right dark>mdi-check-outline</v-icon>
                    </v-btn>
                    <v-btn
                            color="primary"
                            @click="dialogUpdate = false"
                    >
                        Cancel
                        <v-icon right dark>mdi-close-circle-outline</v-icon>
                    </v-btn>
                </v-card-actions>
            </v-card>
        </v-dialog>
    </v-container>
</template>
<script>
  import { mapActions, mapGetters } from 'vuex'

  export default {
    name: 'Clients',

    data: () => ({
      listView: false,
      dialogCreate: false,
      dialogUpdate: false,
      client: null,
      valid: false,
      search: '',
      customTableValue: '',
      headers: [
        { text: 'Name', value: 'name', },
        { text: 'Email', value: 'email', },
        { text: 'IP addresses', value: 'address', },
        { text: 'Tags', value: 'tags', },
        { text: 'Created', value: 'created', sortable: false, },
        { text: 'Updated', value: 'updated', sortable: false, },
        { text: 'Actions', value: 'action', sortable: false, },
      ],
    }),

    computed:{
      ...mapGetters({
        getClientQrcode: 'client/getClientQrcode',
        getClientConfig: 'client/getClientConfig',
        server: 'server/server',
        clients: 'client/clients',
        clientQrcodes: 'client/clientQrcodes',
      }),

      // Available table options
      tableOptions() {
        // Make sure all values are treated as strings for consistency
        return ['auto', 'off'].concat(
          // Include any custom numeric values that might be in the client table field
          typeof this.client?.table === 'number' || 
          (this.client?.table && !isNaN(Number(this.client.table)) && 
           this.client.table !== 'auto' && this.client.table !== 'off') 
            ? [this.client.table.toString()] 
            : []
        );
      },

      // Dynamically determine if endpoint options should be enabled
      showEndpointOptions() {
        return !!this.client.Site2SiteEndpoint; // Returns true if Site2SiteEndpoint is not empty
      },
    },

    mounted () {
      this.readAllClients()
      this.readServer()
    },

    methods: {
      ...mapActions('client', {
        errorClient: 'error',
        readAllClients: 'readAll',
        creatClient: 'create',
        updateClient: 'update',
        deleteClient: 'delete',
        emailClient: 'email',
      }),
      ...mapActions('server', {
        readServer: 'read',
      }),

      startCreate() {
        this.client = {
          name: "",
          email: "",
          enable: true,
          allowedIPs: this.server.allowedips,
          address: this.server.address,
          tags: [],
          ignorePersistentKeepalive: false,
          keepaliveDisabled: false,
          keepaliveInterval: 0,
          useRemoteDNS: false,
          site2site: false,
          lanIPs: [],
          Site2SiteEndpoint: '',
          Site2SiteEndpointListenPort: null,
          Site2SiteEndpointPort: null,
          table: 'auto',
          Site2SiteEndpointOptionsEnabled: false,
        };
        this.dialogCreate = true;
      },

      create(client) {
        // Validate allowed IPs
        if (client.allowedIPs.length < 1) {
          this.errorClient('Please provide at least one valid CIDR address for client allowed IPs');
          return;
        }
        for (let i = 0; i < client.allowedIPs.length; i++) {
          if (this.$isCidr(client.allowedIPs[i]) === 0) {
            this.errorClient('Invalid CIDR detected, please correct before submitting');
            return;
          }
        }

        // Validate address
        if (client.address.length < 1) {
          this.errorClient('Please provide at least one valid CIDR address for client');
          return;
        }
        for (let i = 0; i < client.address.length; i++) {
          if (this.$isCidr(client.address[i]) === 0) {
            this.errorClient('Invalid CIDR detected, please correct before submitting');
            return;
          }
        }

        // Validate keepalive settings
        if (client.ignorePersistentKeepalive && !client.keepaliveDisabled && (!client.keepaliveInterval || client.keepaliveInterval <= 0)) {
          this.errorClient('Keepalive interval must be set when keepalive is enabled');
          return;
        }

        // Validate Site2Site settings
        if (client.site2site) {
          if (client.lanIPs.length === 0) {
            this.errorClient('LAN IPs are required when Site-to-Site is enabled');
            return;
          }
          if (client.Site2SiteEndpointOptionsEnabled) {
            if (client.Site2SiteEndpoint && !client.Site2SiteEndpointListenPort) {
              this.errorClient('Listen port is required if endpoint is set');
              return;
            }
            if (client.Site2SiteEndpoint && client.Site2SiteEndpointPort && client.Site2SiteEndpointPort <= 0) {
              this.errorClient('Public port must be a positive integer');
              return;
            }
          }
        }

        // All good, submit
        this.dialogCreate = false;
        this.creatClient(client);
      },

      remove(client) {
        if(confirm(`Do you really want to delete ${client.name} ?`)){
          this.deleteClient(client)
        }
      },

      email(client) {
        if (!client.email){
          this.errorClient('Client email is not defined')
          return
        }

        if(confirm(`Do you really want to send email to ${client.email} with all configurations ?`)){
          this.emailClient(client)
        }
      },

      startUpdate(client) {
        this.client = {
          ...client,
        };

        // Make sure table is properly handled as string
        if (client.table !== undefined && client.table !== null) {
          // Always ensure table is a string
          this.client.table = client.table.toString();
        } else {
          this.client.table = 'auto'; // Default
        }
        
        this.client.Site2SiteEndpoint = client.Site2SiteEndpoint || '';
        this.client.Site2SiteEndpointListenPort = client.Site2SiteEndpointListenPort || null;
        this.client.Site2SiteEndpointPort = client.Site2SiteEndpointPort || null;
        this.client.Site2SiteEndpointOptionsEnabled = client.Site2SiteEndpointOptionsEnabled || false;

        this.dialogUpdate = true;
      },

      update(client) {
        // Validate allowed IPs
        if (client.allowedIPs.length < 1) {
          this.errorClient('Please provide at least one valid CIDR address for client allowed IPs');
          return;
        }
        for (let i = 0; i < client.allowedIPs.length; i++) {
          if (this.$isCidr(client.allowedIPs[i]) === 0) {
            this.errorClient('Invalid CIDR detected, please correct before submitting');
            return;
          }
        }

        // Validate address
        if (client.address.length < 1) {
          this.errorClient('Please provide at least one valid CIDR address for client');
          return;
        }
        for (let i = 0; i < client.address.length; i++) {
          if (this.$isCidr(client.address[i]) === 0) {
            this.errorClient('Invalid CIDR detected, please correct before submitting');
            return;
          }
        }

        // Validate keepalive settings
        if (client.ignorePersistentKeepalive && !client.keepaliveDisabled && (!client.keepaliveInterval || client.keepaliveInterval <= 0)) {
          this.errorClient('Keepalive interval must be set when keepalive is enabled');
          return;
        }

        // Validate Site2Site settings
        if (client.site2site) {
          if (client.lanIPs.length === 0) {
            this.errorClient('LAN IPs are required when Site-to-Site is enabled');
            return;
          }
          if (client.Site2SiteEndpointOptionsEnabled) {
            if (client.Site2SiteEndpoint && !client.Site2SiteEndpointListenPort) {
              this.errorClient('Listen port is required if endpoint is set');
              return;
            }
            if (client.Site2SiteEndpoint && client.Site2SiteEndpointPort && client.Site2SiteEndpointPort <= 0) {
              this.errorClient('Public port must be a positive integer');
              return;
            }
          }
        }

        // All good, submit
        this.dialogUpdate = false;
        this.updateClient(client);
      },

      forceFileDownload(client){
        let config = this.getClientConfig(client.id)
        if (!config) {
          this.errorClient('Failed to download client config');
          return
        }
        const url = window.URL.createObjectURL(new Blob([config]))
        const link = document.createElement('a')
        link.href = url
        link.setAttribute('download', this.getConfigFileName(client)) //or any other extension
        document.body.appendChild(link)
        link.click()
      },

      getConfigFileName(client){
        let name = client.name.split(' ').join('-');
        // replace special chars
        name = name.replace(/[^a-zA-Z\d_-]+/g, '');
        return name + '.conf';
      },

      handlePersistentKeepaliveChange() {
        if (!this.client.ignorePersistentKeepalive) {
          // If ignorePersistentKeepalive is set to NO
          this.client.keepaliveDisabled = false;
          this.client.keepaliveInterval = 0;
        }
      },
      
      handleKeepaliveDisabledChange() {
        if (this.client.keepaliveDisabled) {
          // If keepaliveDisabled is true, reset interval
          this.client.keepaliveInterval = 0;
        }
      },

      handleSite2SiteChange() {
        if (!this.client.site2site) {
          // If site-to-site is disabled, clear related fields
          this.client.lanIPs = [];
          this.client.Site2SiteEndpoint = '';
          this.client.Site2SiteEndpointListenPort = null;
          this.client.Site2SiteEndpointPort = null;
          this.client.table = 'auto';
          this.client.Site2SiteEndpointOptionsEnabled = false;
        }
      },
      setCustomTableValue() {
        if (this.customTableValue !== '' && !isNaN(this.customTableValue)) {
          // Convert to integer but store as string to avoid type issues with the backend
          const tableValue = parseInt(this.customTableValue, 10);
          
          // Set the value directly to the client object, but convert to string
          this.client.table = tableValue.toString();
          
          // Clear the input field
          this.customTableValue = '';
        }
      },

      handleEnableEndpointOptionsChange() {
        if (!this.client.Site2SiteEndpointOptionsEnabled) {
          // If the switch is set to NO, reset the related fields
          this.client.Site2SiteEndpoint = '';
          this.client.Site2SiteEndpointListenPort = null;
          this.client.Site2SiteEndpointPort = null;
        }
      },

      validatePortInput(event, fieldName) {
        // If the field is empty or not a number, set it to null
        if (event.target.value === '' || isNaN(event.target.value)) {
          this.client[fieldName] = null;
        } else {
          // Otherwise, ensure it's a number
          this.client[fieldName] = parseInt(event.target.value, 10);
        }
      },
    }
  };
</script>
