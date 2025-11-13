<template>
  <div class="profile-shell">
    <section class="profile-hero">
      <div class="hero-primary">
        <img src="https://placehold.co/140x140/161832/fff?text=ME" alt="profile avatar" />
        <div>
          <p class="nickname">@neonpilot</p>
          <h1>Marina Pulse</h1>
          <p class="about">
            {{ aboutText }}
          </p>
          <div class="hero-stats">
            <div
              v-for="stat in stats"
              :key="stat.label"
              :class="['stat-block', { clickable: isPanelStat(stat.label) }]"
              @click="handleStatInteraction(stat.label)"
              @keydown.enter.prevent="handleStatInteraction(stat.label)"
              @keydown.space.prevent="handleStatInteraction(stat.label)"
              :role="isPanelStat(stat.label) ? 'button' : undefined"
              :tabindex="isPanelStat(stat.label) ? 0 : undefined"
            >
              <span>{{ stat.value }}</span>
              <small>{{ stat.label }}</small>
            </div>
          </div>
        </div>
      </div>
        <div class="hero-controls">
          <button class="privacy-toggle" @click="togglePrivacy">
            <span :class="['status-dot', isPrivate ? 'dot-private' : 'dot-public']"></span>
            {{ isPrivate ? 'Private mode' : 'Public mode' }}
          </button>
          <p class="privacy-copy">
          {{
            isPrivate
              ? 'Only accepted followers can see your content and profile fields.'
              : 'Everyone can view your highlights, posts, and profile fields.'
          }}
        </p>
        <div class="hero-buttons">
          <button class="ghost" @click="emit('back')">Back to feed</button>
          <button class="ghost" @click="toggleInfoMode">{{ infoMode ? 'Hide info' : 'Info' }}</button>
          <button class="cta">Share profile</button>
        </div>
      </div>
    </section>

    <Transition name="info-fade">
      <section v-if="infoMode" class="info-grid">
        <article v-for="section in infoSections" :key="section.key">
          <header>
            <div>
              <small>{{ section.label }}</small>
              <h3>{{ section.value }}</h3>
            </div>
            <div class="info-actions">
              <button
                class="icon-btn"
                @click="toggleVisibility(section.key)"
                :title="section.visible ? 'Hide' : 'Show'"
              >
                <span :class="['icon-eye', section.visible ? 'visible' : 'hidden']"></span>
              </button>
              <button class="ghost mini" @click="openEditor(section.key)">Edit</button>
            </div>
          </header>
          <span :class="['visibility-pill', section.visible ? 'pill-visible' : 'pill-hidden']">
            {{ section.visible ? 'Displayed' : 'Hidden' }}
          </span>
        </article>
      </section>
    </Transition>

    <section class="activity-panel">
      <div class="activity-body">
        <div class="post-stack">
          <article v-for="post in mockPosts" :key="post.id" class="post-card">
            <header>
              <strong>{{ post.title }}</strong>
              <small>{{ post.time }}</small>
            </header>
            <p>{{ post.body }}</p>
          </article>
        </div>
      </div>
    </section>

    <div v-if="infoMode && editingSection" class="edit-drawer">
      <div class="drawer-body">
        <header>
          <h3>Edit {{ currentEdit?.label }}</h3>
          <button class="icon-btn" @click="closeEditor">
            ✕
          </button>
        </header>
        <p>
          This is a placeholder editor panel. Replace with a real form once backend wiring is ready.
        </p>
        <label>
          <span>{{ currentEdit?.label }}</span>
          <component
            :is="currentEdit?.key === 'about' ? 'textarea' : 'input'"
            v-model="draftValue"
            :rows="currentEdit?.key === 'about' ? 5 : undefined"
            class="editor-field"
          />
        </label>
        <div class="drawer-actions">
          <button class="ghost" @click="closeEditor">Cancel</button>
          <button class="cta" @click="saveEditor">Save</button>
        </div>
      </div>
    </div>

    <Transition name="side-panel">
      <aside v-if="followersPanelOpen" class="side-panel followers-panel">
        <header>
          <div>
            <small>Followers</small>
            <h3>{{ mockFollowers.length }} total</h3>
          </div>
          <button class="icon-btn close-btn" aria-label="Close followers list" @click="closeFollowersPanel">
            ✕
          </button>
        </header>
        <label class="panel-search">
          <span class="sr-only">Search followers</span>
          <input
            v-model="followerSearch"
            type="search"
            placeholder="Search followers..."
            autocomplete="off"
          />
        </label>
        <div class="panel-list">
          <article v-for="user in filteredFollowers" :key="user.handle" class="panel-row">
            <img :src="user.avatar" :alt="`${user.name} avatar`" />
            <div>
              <strong>{{ user.name }}</strong>
              <small>{{ user.handle }}</small>
            </div>
            <button class="ghost mini">View profile</button>
          </article>
          <p v-if="!filteredFollowers.length" class="empty-state">
            No followers match your search.
          </p>
        </div>
      </aside>
    </Transition>

    <Transition name="side-panel">
      <aside v-if="followingPanelOpen" class="side-panel following-panel">
        <header>
          <div>
            <small>Following</small>
            <h3>{{ mockFollowing.length }} accounts</h3>
          </div>
          <button class="icon-btn close-btn" aria-label="Close following list" @click="closeFollowingPanel">
            ✕
          </button>
        </header>
        <label class="panel-search">
          <span class="sr-only">Search following</span>
          <input
            v-model="followingSearch"
            type="search"
            placeholder="Search following..."
            autocomplete="off"
          />
        </label>
        <div class="panel-list">
          <article v-for="user in filteredFollowing" :key="user.handle" class="panel-row">
            <img :src="user.avatar" :alt="`${user.name} avatar`" />
            <div>
              <strong>{{ user.name }}</strong>
              <small>{{ user.handle }}</small>
            </div>
            <button class="ghost mini">Message</button>
          </article>
          <p v-if="!filteredFollowing.length" class="empty-state">
            No accounts match your search.
          </p>
        </div>
      </aside>
    </Transition>

    <Transition name="side-panel">
      <aside v-if="groupsPanelOpen" class="side-panel groups-panel">
        <header>
          <div>
            <small>Groups</small>
            <h3>{{ mockGroups.length }} joined</h3>
          </div>
          <button class="icon-btn close-btn" aria-label="Close groups list" @click="closeGroupsPanel">
            ✕
          </button>
        </header>
        <label class="panel-search">
          <span class="sr-only">Search groups</span>
          <input
            v-model="groupSearch"
            type="search"
            placeholder="Search groups..."
            autocomplete="off"
          />
        </label>
        <div class="panel-list groups-list">
          <article v-for="group in filteredGroups" :key="group.title" class="panel-row group-row">
            <div class="group-row__body">
              <strong>{{ group.title }}</strong>
              <small>{{ group.members }} members</small>
              <p>{{ group.desc }}</p>
            </div>
            <button class="ghost mini">Open</button>
          </article>
          <p v-if="!filteredGroups.length" class="empty-state">
            No groups match your search.
          </p>
        </div>
      </aside>
    </Transition>

    <Transition name="side-panel">
      <aside v-if="eventsPanelOpen" class="side-panel events-panel">
        <header>
          <div>
            <small>Events</small>
            <h3>{{ mockEvents.length }} upcoming</h3>
          </div>
          <button class="icon-btn close-btn" aria-label="Close events list" @click="closeEventsPanel">
            ✕
          </button>
        </header>
        <label class="panel-search">
          <span class="sr-only">Search events</span>
          <input
            v-model="eventSearch"
            type="search"
            placeholder="Search events..."
            autocomplete="off"
          />
        </label>
        <div class="panel-list events-list">
          <article v-for="event in filteredEvents" :key="event.title" class="panel-row event-row">
            <div class="event-row__meta">
              <strong>{{ event.title }}</strong>
              <small>{{ event.date }}</small>
            </div>
            <p>{{ event.desc }}</p>
            <span class="event-location">{{ event.location }}</span>
            <button class="ghost mini">Details</button>
          </article>
          <p v-if="!filteredEvents.length" class="empty-state">
            No events match your search.
          </p>
        </div>
      </aside>
    </Transition>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue';

const emit = defineEmits(['back']);

const isPrivate = ref(false);
const editingSection = ref('');
const infoMode = ref(false);
const draftValue = ref('');
const followersPanelOpen = ref(false);
const followingPanelOpen = ref(false);
const groupsPanelOpen = ref(false);
const eventsPanelOpen = ref(false);
const followerSearch = ref('');
const followingSearch = ref('');
const groupSearch = ref('');
const eventSearch = ref('');

const stats = [
  { label: 'Followers', value: '1.3K' },
  { label: 'Following', value: '402' },
  { label: 'Groups', value: '8' },
  { label: 'Events', value: '3' },
];

const panelStatLabels = ['Followers', 'Following', 'Groups', 'Events'];

const infoSections = ref([
  { key: 'firstName', label: 'First Name', value: 'Marina', visible: true },
  { key: 'lastName', label: 'Last Name', value: 'Pulse', visible: true },
  { key: 'dob', label: 'Date of Birth', value: '1994-08-17', visible: false },
  { key: 'nickname', label: 'Nickname', value: 'Neon Pilot', visible: true },
  { key: 'about', label: 'About Me', value: 'Collecting retro synths & designing social UX.', visible: true },
]);

const aboutText = computed(() => infoSections.value.find((section) => section.key === 'about')?.value ?? '');

const mockPosts = [
  { id: 'p1', title: 'Private BETA', time: '2h ago', body: 'Rolling out the “almost private” feed filter to my inner circle.' },
  { id: 'p2', title: 'Group Event', time: '1d ago', body: 'Synthwave Creators meetup this Saturday. RSVP if you are going.' },
];

const mockFollowers = [
  { name: 'Nova Flux', handle: '@nova', avatar: 'https://placehold.co/64x64/15162a/fff?text=NF' },
  { name: 'Echo Lane', handle: '@echo', avatar: 'https://placehold.co/64x64/20223c/fff?text=EL' },
  { name: 'Lumen Rae', handle: '@lumen', avatar: 'https://placehold.co/64x64/181a33/fff?text=LR' },
];

const filteredFollowers = computed(() => {
  const term = followerSearch.value.trim().toLowerCase();
  if (!term) return mockFollowers;
  return mockFollowers.filter((user) => {
    return [user.name, user.handle].some((field) => field.toLowerCase().includes(term));
  });
});

const filteredFollowing = computed(() => {
  const term = followingSearch.value.trim().toLowerCase();
  if (!term) return mockFollowing;
  return mockFollowing.filter((user) => {
    return [user.name, user.handle].some((field) => field.toLowerCase().includes(term));
  });
});

const filteredGroups = computed(() => {
  const term = groupSearch.value.trim().toLowerCase();
  if (!term) return mockGroups;
  return mockGroups.filter((group) => {
    return [group.title, group.desc].some((field) => field.toLowerCase().includes(term));
  });
});

const filteredEvents = computed(() => {
  const term = eventSearch.value.trim().toLowerCase();
  if (!term) return mockEvents;
  return mockEvents.filter((event) => {
    return [event.title, event.desc, event.location].some((field) => field.toLowerCase().includes(term));
  });
});

const mockFollowing = [
  { name: 'Glitch Bloom', handle: '@glitch', avatar: 'https://placehold.co/64x64/161832/fff?text=GB' },
  { name: 'Circuit Club', handle: '@circuit', avatar: 'https://placehold.co/64x64/101126/fff?text=CC' },
];

const mockGroups = [
  { title: 'Synthwave Creators', members: 128, desc: 'Designers pushing neon themed UX.' },
  { title: 'Hyperlink Hub', members: 82, desc: 'Invite-only crew testing privacy tools.' },
];

const mockEvents = [
  {
    title: 'Neon Nights Meetup',
    date: 'Fri, Mar 22',
    location: 'Arcade District',
    desc: 'Showcase your latest glow UI concepts and retro synth sets.',
  },
  {
    title: 'Pulse Design Sprint',
    date: 'Tue, Mar 26',
    location: 'Virtual • Holo Conference',
    desc: 'Collaborate on privacy-first social flows and prototypes.',
  },
  {
    title: 'Synthwave Creator Jam',
    date: 'Sun, Mar 31',
    location: 'Downtown Studio 07',
    desc: 'Live jam session with surprise guest VJs.',
  },
];

const currentEdit = computed(() => infoSections.value.find((section) => section.key === editingSection.value));

function togglePrivacy() {
  isPrivate.value = !isPrivate.value;
}

function toggleVisibility(key) {
  infoSections.value = infoSections.value.map((section) =>
    section.key === key ? { ...section, visible: !section.visible } : section
  );
}

function openEditor(key) {
  if (!infoMode.value) return;
  editingSection.value = key;
  draftValue.value = infoSections.value.find((section) => section.key === key)?.value ?? '';
}

function closeEditor() {
  editingSection.value = '';
  draftValue.value = '';
}

function saveEditor() {
  infoSections.value = infoSections.value.map((section) =>
    section.key === editingSection.value ? { ...section, value: draftValue.value } : section
  );
  closeEditor();
}

function toggleInfoMode() {
  infoMode.value = !infoMode.value;
  if (!infoMode.value) {
    closeEditor();
  }
}

function isPanelStat(label) {
  return panelStatLabels.includes(label);
}

function handleStatInteraction(label) {
  if (!isPanelStat(label)) return;
  if (label === 'Followers') {
    followersPanelOpen.value ? closeFollowersPanel() : openFollowersPanel();
  } else if (label === 'Following') {
    followingPanelOpen.value ? closeFollowingPanel() : openFollowingPanel();
  } else if (label === 'Groups') {
    groupsPanelOpen.value ? closeGroupsPanel() : openGroupsPanel();
  } else if (label === 'Events') {
    eventsPanelOpen.value ? closeEventsPanel() : openEventsPanel();
  }
}

function closeAllPanels() {
  followersPanelOpen.value = false;
  followingPanelOpen.value = false;
  groupsPanelOpen.value = false;
  eventsPanelOpen.value = false;
}

function openFollowersPanel() {
  closeAllPanels();
  followersPanelOpen.value = true;
}

function openFollowingPanel() {
  closeAllPanels();
  followingPanelOpen.value = true;
}

function openGroupsPanel() {
  closeAllPanels();
  groupsPanelOpen.value = true;
}

function openEventsPanel() {
  closeAllPanels();
  eventsPanelOpen.value = true;
}

function closeFollowersPanel() {
  followersPanelOpen.value = false;
  followerSearch.value = '';
}

function closeFollowingPanel() {
  followingPanelOpen.value = false;
  followingSearch.value = '';
}

function closeGroupsPanel() {
  groupsPanelOpen.value = false;
  groupSearch.value = '';
}

function closeEventsPanel() {
  eventsPanelOpen.value = false;
  eventSearch.value = '';
}
</script>

<style scoped>
.profile-shell {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.profile-hero {
  border-radius: 1.75rem;
  padding: clamp(1.5rem, 4vw, 2.5rem);
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: linear-gradient(120deg, rgba(0, 247, 255, 0.1), rgba(255, 0, 230, 0.08));
  display: flex;
  flex-wrap: wrap;
  gap: 1.5rem;
  justify-content: space-between;
}

.hero-primary {
  display: flex;
  gap: 1.5rem;
  align-items: center;
  min-width: 260px;
}

.hero-primary img {
  width: 140px;
  height: 140px;
  border-radius: 1.5rem;
  border: 2px solid rgba(255, 255, 255, 0.15);
  object-fit: cover;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.45);
}

.hero-primary h1 {
  margin: 0.35rem 0;
  font-size: clamp(1.8rem, 4vw, 2.4rem);
}

.nickname {
  text-transform: uppercase;
  letter-spacing: 0.2em;
  font-size: 0.7rem;
  color: var(--neon-cyan);
  margin: 0;
}

.about {
  color: var(--text-muted);
  max-width: 36ch;
}

.hero-stats {
  display: flex;
  gap: 1.5rem;
  margin-top: 1rem;
}

.stat-block {
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
  transition: color 0.2s ease;
}

.stat-block.clickable {
  cursor: pointer;
}

.stat-block.clickable:hover span,
.stat-block.clickable:focus-visible span {
  color: var(--neon-cyan);
}

.hero-stats span {
  font-size: 1.3rem;
  font-weight: 600;
}

.hero-stats small {
  color: var(--text-muted);
  letter-spacing: 0.05em;
}

.hero-controls {
  max-width: 320px;
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}

.privacy-toggle {
  display: inline-flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.65rem 1rem;
  border-radius: 999px;
  border: 1px solid var(--border-glow);
  background: rgba(3, 5, 12, 0.6);
  color: inherit;
  cursor: pointer;
}

.status-dot {
  width: 0.75rem;
  height: 0.75rem;
  border-radius: 999px;
  box-shadow: 0 0 10px currentColor;
}

.dot-public {
  color: var(--neon-cyan);
  background: var(--neon-cyan);
}

.dot-private {
  color: var(--neon-pink);
  background: var(--neon-pink);
}

.hero-buttons {
  display: flex;
  gap: 0.75rem;
}

.cta {
  border: none;
  border-radius: 999px;
  padding: 0.6rem 1.4rem;
  background: linear-gradient(120deg, var(--neon-cyan), var(--neon-pink));
  color: #05060d;
  font-weight: 600;
  cursor: pointer;
  box-shadow: 0 10px 25px rgba(255, 0, 230, 0.3);
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
}

.info-fade-enter-active,
.info-fade-leave-active {
  transition: opacity 0.35s ease, transform 0.35s ease;
}

.info-fade-enter-from,
.info-fade-leave-to {
  opacity: 0;
  transform: translateY(16px) scale(0.98);
}

.info-grid article {
  padding: 1.25rem;
  border-radius: 1.25rem;
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(6, 8, 18, 0.8);
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.info-grid h3 {
  margin: 0.35rem 0 0;
}

.info-actions {
  display: flex;
  gap: 0.35rem;
  align-items: center;
}

.icon-btn {
  width: 2rem;
  height: 2rem;
  border-radius: 0.75rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(8, 10, 22, 0.7);
  display: grid;
  place-items: center;
  cursor: pointer;
}

.ghost {
  background: transparent;
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: inherit;
  border-radius: 999px;
  padding: 0.45rem 1.1rem;
  cursor: pointer;
}

.ghost.mini {
  padding: 0.3rem 0.9rem;
}

.visibility-pill {
  width: fit-content;
  padding: 0.2rem 0.75rem;
  border-radius: 999px;
  font-size: 0.7rem;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.pill-visible {
  background: rgba(0, 247, 255, 0.18);
  color: var(--neon-cyan);
}

.pill-hidden {
  background: rgba(255, 0, 230, 0.18);
  color: var(--neon-pink);
}

.activity-panel {
  border-radius: 1.5rem;
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(5, 6, 13, 0.85);
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.activity-body .post-card {
  padding: 1.25rem;
  border-radius: 1.25rem;
  border: 1px solid rgba(255, 255, 255, 0.07);
  background: rgba(10, 12, 22, 0.8);
  margin-bottom: 1rem;
}

.edit-drawer {
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  width: min(360px, 80vw);
  background: rgba(5, 6, 13, 0.7);
  border-left: 1px solid rgba(255, 255, 255, 0.05);
  box-shadow: -12px 0 30px rgba(0, 0, 0, 0.25);
  backdrop-filter: blur(16px);
  padding: 1.75rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  height: 100%;
}

.drawer-body {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  height: 100%;
}

.drawer-body .editor-field {
  width: 100%;
  padding: 0.65rem;
  border-radius: 0.75rem;
  border: 1px solid rgba(255, 255, 255, 0.2);
  background: rgba(3, 4, 10, 0.6);
  color: inherit;
}

.drawer-body textarea.editor-field {
  min-height: 140px;
  resize: vertical;
}

.drawer-actions {
  margin-top: auto;
  display: flex;
  gap: 0.75rem;
  justify-content: flex-end;
}

.icon-eye {
  display: inline-block;
  width: 1.15rem;
  height: 1.15rem;
  mask: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="white" d="M12 5c-7.633 0-11 7-11 7s3.367 7 11 7 11-7 11-7-3.367-7-11-7zm0 11a4 4 0 1 1 .001-8.001A4 4 0 0 1 12 16zm0-6a2 2 0 1 0 .001 3.999A2 2 0 0 0 12 10z"/></svg>')
    center/contain no-repeat;
}

.icon-eye.visible {
  background: var(--neon-cyan);
}

.icon-eye.hidden {
  background: var(--neon-pink);
  mask: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="white" d="M12 5c-1.48 0-2.834.249-4.063.642L5.707 3.413 4.293 4.827l14.88 14.88 1.414-1.414-3.07-3.07C20.181 12.9 23 12 23 12s-3.367-7-11-7zm-4.95 3.536a14.63 14.63 0 0 1 4.95-.82c4.904 0 7.875 2.397 9.163 3.864-.41.454-1.17 1.17-2.24 1.907l-2.78-2.78a5 5 0 0 0-8.308-2.171l-0.785-.785zm8.528 8.528A14.63 14.63 0 0 1 12 18.18c-4.904 0-7.875-2.397-9.163-3.864a27.09 27.09 0 0 1 2.167-1.774l-1.56-1.56L2 12s3.367 7 11 7c1.683 0 3.193-.285 4.578-.936z"/></svg>')
    center/contain no-repeat;
}

.side-panel {
  position: fixed;
  top: 0;
  bottom: 0;
  left: 0;
  width: min(360px, 85vw);
  padding: 1.75rem 1.5rem;
  background: rgba(5, 6, 13, 0.92);
  border-right: 1px solid rgba(255, 255, 255, 0.08);
  box-shadow: 10px 0 30px rgba(0, 0, 0, 0.35);
  backdrop-filter: blur(18px);
  display: flex;
  flex-direction: column;
  gap: 1rem;
  z-index: 30;
}

.side-panel header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.side-panel header small {
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--text-muted);
}

.close-btn {
  border-radius: 999px;
  width: 2.25rem;
  height: 2.25rem;
  border: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(0, 0, 0, 0.35);
  color: rgba(255, 255, 255, 0.85);
  font-size: 1.1rem;
  transition: border-color 0.2s ease, color 0.2s ease, background 0.2s ease;
}

.close-btn:hover,
.close-btn:focus-visible {
  border-color: rgba(0, 247, 255, 0.6);
  color: var(--neon-cyan);
  background: rgba(0, 0, 0, 0.55);
}

.panel-search {
  width: 100%;
}

.panel-search input {
  width: 100%;
  padding: 0.65rem 0.85rem;
  border-radius: 0.9rem;
  border: 1px solid rgba(255, 255, 255, 0.2);
  background: rgba(8, 9, 20, 0.8);
  color: inherit;
}

.panel-list {
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.panel-row {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  padding: 0.9rem;
  border-radius: 1rem;
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(7, 8, 18, 0.85);
}

.panel-row img {
  width: 44px;
  height: 44px;
  border-radius: 50%;
}

.groups-list .panel-row {
  align-items: flex-start;
}

.group-row__body {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.group-row__body p {
  margin: 0.1rem 0 0;
  color: var(--text-muted);
  font-size: 0.85rem;
}

.events-list .panel-row {
  flex-direction: column;
  align-items: flex-start;
  gap: 0.4rem;
}

.event-row__meta {
  display: flex;
  justify-content: space-between;
  width: 100%;
  gap: 0.5rem;
}

.event-row__meta small {
  color: var(--text-muted);
}

.event-row p {
  margin: 0;
  color: var(--text-muted);
}

.event-location {
  font-size: 0.85rem;
  color: var(--neon-cyan);
}

.empty-state {
  text-align: center;
  color: var(--text-muted);
  padding: 1rem 0;
}

.side-panel-enter-active,
.side-panel-leave-active {
  transition: transform 0.35s ease, opacity 0.3s ease;
}

.side-panel-enter-from,
.side-panel-leave-to {
  transform: translateX(-100%);
  opacity: 0;
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  border: 0;
}

@media (max-width: 768px) {
  .hero-primary {
    flex-direction: column;
    align-items: flex-start;
  }

  .hero-stats {
    flex-wrap: wrap;
  }

  .hero-controls {
    width: 100%;
  }
}
</style>
