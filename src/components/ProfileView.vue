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
            <div v-for="stat in stats" :key="stat.label">
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
      <div class="activity-tabs">
        <button
          v-for="tab in activityTabs"
          :key="tab.id"
          :class="['tab-btn', { active: activeTab === tab.id }]"
          @click="activeTab = tab.id"
        >
          <span :class="['icon', tab.icon]"></span>
          {{ tab.label }}
        </button>
      </div>

      <div class="activity-body">
        <div v-if="activeTab === 'posts'" class="post-stack">
          <article v-for="post in mockPosts" :key="post.id" class="post-card">
            <header>
              <strong>{{ post.title }}</strong>
              <small>{{ post.time }}</small>
            </header>
            <p>{{ post.body }}</p>
          </article>
        </div>

        <div v-else-if="activeTab === 'followers'" class="card-grid">
          <article v-for="user in mockFollowers" :key="user.handle" class="follow-card">
            <img :src="user.avatar" alt="" />
            <div>
              <strong>{{ user.name }}</strong>
              <small>{{ user.handle }}</small>
            </div>
            <button class="ghost mini">Remove</button>
          </article>
        </div>

        <div v-else-if="activeTab === 'following'" class="card-grid">
          <article v-for="user in mockFollowing" :key="user.handle" class="follow-card">
            <img :src="user.avatar" alt="" />
            <div>
              <strong>{{ user.name }}</strong>
              <small>{{ user.handle }}</small>
            </div>
            <button class="ghost mini">Message</button>
          </article>
        </div>

        <div v-else class="card-grid groups">
          <article v-for="group in mockGroups" :key="group.title">
            <header>
              <strong>{{ group.title }}</strong>
              <small>{{ group.members }} members</small>
            </header>
            <p>{{ group.desc }}</p>
            <button class="ghost mini">View group</button>
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
  </div>
</template>

<script setup>
import { computed, ref } from 'vue';

const emit = defineEmits(['back']);

const isPrivate = ref(false);
const editingSection = ref('');
const infoMode = ref(false);
const draftValue = ref('');

const stats = [
  { label: 'Followers', value: '1.3K' },
  { label: 'Following', value: '402' },
  { label: 'Groups', value: '8' },
  { label: 'Events', value: '3' },
];

const infoSections = ref([
  { key: 'firstName', label: 'First Name', value: 'Marina', visible: true },
  { key: 'lastName', label: 'Last Name', value: 'Pulse', visible: true },
  { key: 'dob', label: 'Date of Birth', value: '1994-08-17', visible: false },
  { key: 'nickname', label: 'Nickname', value: 'Neon Pilot', visible: true },
  { key: 'about', label: 'About Me', value: 'Collecting retro synths & designing social UX.', visible: true },
]);

const aboutText = computed(() => infoSections.value.find((section) => section.key === 'about')?.value ?? '');

const activityTabs = [
  { id: 'posts', label: 'Posts', icon: 'icon-posts' },
  { id: 'followers', label: 'Followers', icon: 'icon-followers' },
  { id: 'following', label: 'Following', icon: 'icon-following' },
  { id: 'groups', label: 'Groups', icon: 'icon-groups' },
];

const activeTab = ref('posts');

const mockPosts = [
  { id: 'p1', title: 'Private BETA', time: '2h ago', body: 'Rolling out the “almost private” feed filter to my inner circle.' },
  { id: 'p2', title: 'Group Event', time: '1d ago', body: 'Synthwave Creators meetup this Saturday. RSVP if you are going.' },
];

const mockFollowers = [
  { name: 'Nova Flux', handle: '@nova', avatar: 'https://placehold.co/64x64/15162a/fff?text=NF' },
  { name: 'Echo Lane', handle: '@echo', avatar: 'https://placehold.co/64x64/20223c/fff?text=EL' },
  { name: 'Lumen Rae', handle: '@lumen', avatar: 'https://placehold.co/64x64/181a33/fff?text=LR' },
];

const mockFollowing = [
  { name: 'Glitch Bloom', handle: '@glitch', avatar: 'https://placehold.co/64x64/161832/fff?text=GB' },
  { name: 'Circuit Club', handle: '@circuit', avatar: 'https://placehold.co/64x64/101126/fff?text=CC' },
];

const mockGroups = [
  { title: 'Synthwave Creators', members: 128, desc: 'Designers pushing neon themed UX.' },
  { title: 'Hyperlink Hub', members: 82, desc: 'Invite-only crew testing privacy tools.' },
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

.activity-tabs {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 0.75rem;
}

.tab-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  padding: 0.75rem 1rem;
  border-radius: 1rem;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(8, 9, 20, 0.8);
  cursor: pointer;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.tab-btn.active {
  border-color: var(--border-glow);
  box-shadow: 0 0 18px rgba(0, 247, 255, 0.2);
}

.activity-body .post-card {
  padding: 1.25rem;
  border-radius: 1.25rem;
  border: 1px solid rgba(255, 255, 255, 0.07);
  background: rgba(10, 12, 22, 0.8);
  margin-bottom: 1rem;
}

.card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1rem;
}

.follow-card {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem;
  border-radius: 1rem;
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(7, 9, 18, 0.75);
}

.follow-card img {
  width: 48px;
  height: 48px;
  border-radius: 999px;
}

.groups article {
  padding: 1.2rem;
  border-radius: 1.25rem;
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(9, 10, 20, 0.8);
  display: flex;
  flex-direction: column;
  gap: 0.7rem;
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

.icon-posts,
.icon-followers,
.icon-following,
.icon-groups {
  width: 1rem;
  height: 1rem;
  background: var(--neon-cyan);
  mask: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="white" d="M4 6h16v2H4zm0 5h10v2H4zm0 5h16v2H4z"/></svg>')
    center/contain no-repeat;
}

.icon-followers {
  mask: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="white" d="M16 11c1.657 0 3-1.343 3-3S17.657 5 16 5s-3 1.343-3 3 1.343 3 3 3zm-8 0c1.657 0 3-1.343 3-3S9.657 5 8 5 5 6.343 5 8s1.343 3 3 3zm0 2c-2.671 0-8 1.337-8 4v2h12v-2c0-2.663-5.329-4-8-4zm8 0c-.312 0-.663.019-1.031.052C16.964 13.827 19 15.041 19 17v2h5v-2c0-2.663-5.329-4-8-4z"/></svg>')
    center/contain no-repeat;
}

.icon-following {
  mask: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="white" d="M12 2a5 5 0 0 1 5 5 5 5 0 0 1-1.09 3.134A7 7 0 0 1 19 17v3h-2v-3a5 5 0 0 0-10 0v3H5v-3a7 7 0 0 1 3.09-6.866A5 5 0 0 1 12 2z"/></svg>')
    center/contain no-repeat;
}

.icon-groups {
  mask: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="white" d="M12 5a4 4 0 0 1 4 4 4 4 0 0 1-1.11 2.767A6 6 0 0 1 19 17v3h-2v-3a4 4 0 0 0-8 0v3H7v-3a6 6 0 0 1 4.11-5.233A4 4 0 0 1 12 5zm-6-1a3 3 0 0 1 3 3 3 3 0 0 1-.274 1.26A5 5 0 0 1 11 13v2H1v-2a5 5 0 0 1 2.274-4.74A3 3 0 0 1 6 4zm12 0a3 3 0 0 1 3 3 3 3 0 0 1-.274 1.26A5 5 0 0 1 23 13v2h-8v-2a5 5 0 0 1 2.274-4.74A3 3 0 0 1 18 4z"/></svg>')
    center/contain no-repeat;
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
