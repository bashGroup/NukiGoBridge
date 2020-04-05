<template>
  <v-app id="inspire">
    <v-navigation-drawer v-model="drawer" app clipped>
      <v-list-item>
        <v-list-item-content>
          <v-list-item-title class="title">Your Locks</v-list-item-title>
        </v-list-item-content>
      </v-list-item>
      <v-divider></v-divider>
      <v-list dense>
        <template v-for="lock in locks">
          <v-list-item link :key="lock.id">
            <v-list-item-content>
              <v-list-item-title>{{ lock.name }}</v-list-item-title>
              <v-list-item-subtitle>
                <span v-if="lock.locked">locked</span>
                <span v-else>unlocked</span> |
                <span v-if="lock.doorSensorState">door open</span>
                <span v-else>door closed</span>
              </v-list-item-subtitle>
            </v-list-item-content>
            <v-list-item-action>
              <v-list-item-action-text>{{ lock.timestamp }}</v-list-item-action-text>
              <v-row>
                <v-tooltip top v-if="!lock.doorSensorState">
                  <span>Unlatch and open the door</span>
                  <template v-slot:activator="{ on }">
                    <v-btn icon v-on="on">
                      <v-icon>mdi-door-open</v-icon>
                    </v-btn>
                  </template>
                </v-tooltip>
                <v-tooltip top v-if="!lock.doorSensorState && lock.locked">
                  <span>Unlock the door</span>
                  <template v-slot:activator="{ on }">
                    <v-btn icon v-on="on">
                      <v-icon>mdi-lock</v-icon>
                    </v-btn>
                  </template>
                </v-tooltip>

                <v-tooltip top v-if="!lock.doorSensorState && !lock.locked">
                  <span>Lock the door</span>
                  <template v-slot:activator="{ on }">
                    <v-btn icon v-on="on">
                      <v-icon>mdi-lock-open</v-icon>
                    </v-btn>
                  </template>
                </v-tooltip>
              </v-row>
            </v-list-item-action>
          </v-list-item>
        </template>
      </v-list>
      <template v-slot:append>
        <v-divider></v-divider>
        <v-list-item link>
          <v-list-item-action>
            <v-icon>mdi-cog</v-icon>
          </v-list-item-action>
          <v-list-item-content>
            <v-list-item-title>Settings</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item link>
          <v-list-item-action>
            <v-icon>mdi-help</v-icon>
          </v-list-item-action>
          <v-list-item-content>
            <v-list-item-title>Help</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
      </template>
    </v-navigation-drawer>

    <v-app-bar :clipped-left="$vuetify.breakpoint.lgAndUp" app color="blue darken-3" dark>
      <v-app-bar-nav-icon @click.stop="drawer = !drawer" />
      <v-toolbar-title style="width: 300px" class="ml-0 pl-4">
        <span class="hidden-sm-and-down">NukiGoBridge</span>
      </v-toolbar-title>
      <v-spacer />
    </v-app-bar>
    <v-content>
      <v-card>
        <v-tabs icons-and-text fixed-tabs >
          <v-tab href="#tab-1">
            <v-icon>mdi-information</v-icon>State
          </v-tab>

          <v-tab href="#tab-2">
            <v-icon>mdi-timeline-clock</v-icon>History
          </v-tab>

          <v-tab href="#tab-3">
            <v-icon>mdi-cog</v-icon>Settings
          </v-tab>
          <v-tab-item v-for="i in 3" :key="i" :value="'tab-' + i">
            <v-card flat max-width="600">
              <v-timeline>
                <v-timeline-item v-for="entry in history" :key="entry.index">
                  <span slot="opposite">2020-04-05T15:18:00Z</span>
                  <v-card class="elevation-2">
                    <v-card-title class="headline"><span v-if="entry.type==6 && entry.details.doorSensor==1">Door closed</span><span v-if="entry.type==6 && entry.details.doorSensor==0">Door opened</span></v-card-title>
                  </v-card>
                </v-timeline-item>
              </v-timeline>
            </v-card>
          </v-tab-item>
        </v-tabs>
      </v-card>
    </v-content>
    <v-btn bottom color="pink" dark fab fixed right @click="dialog = !dialog">
      <v-icon>mdi-plus</v-icon>
    </v-btn>
  </v-app>
</template>

<script>
export default {
  props: {
    source: String
  },
  data: () => ({
    locks: [
      {
        id: 4323213,
        name: "Front Door",
        locked: false,
        doorSensorState: true,
        timestamp: "yesterday"
      },
      {
        id: 424231,
        name: "Back Door",
        locked: true,
        doorSensorState: false,
        timestamp: "2h"
      },
      {
        id: 424231,
        name: "Garage",
        locked: false,
        doorSensorState: false,
        timestamp: "5m"
      }
    ],
    history: [
      {
        index: 1161,
        timestamp: "2020-04-05T15:18:00Z",
        authId: 4294967295,
        name: "",
        type: 6,
        details: {
          doorSensor: 1
        }
      },
      {
        index: 1160,
        timestamp: "2020-04-05T15:17:27Z",
        authId: 4294967295,
        name: "",
        type: 6,
        details: {
          doorSensor: 0
        }
      },
      {
        index: 1159,
        timestamp: "2020-04-05T13:30:04Z",
        authId: 4294967295,
        name: "",
        type: 6,
        details: {
          doorSensor: 1
        }
      },
      {
        index: 1158,
        timestamp: "2020-04-05T13:29:32Z",
        authId: 4294967295,
        name: "",
        type: 6,
        details: {
          doorSensor: 0
        }
      },
      {
        index: 1157,
        timestamp: "2020-04-05T13:27:42Z",
        authId: 4294967295,
        name: "",
        type: 6,
        details: {
          doorSensor: 1
        }
      }
    ]
  })
};
</script>